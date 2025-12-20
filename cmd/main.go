package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"ffmpeg/internal"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"ffmpeg/api"
)

func main() {
	http.HandleFunc("/v1/process", handleProcess)
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		internal.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	defer r.Body.Close()

	var req api.ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	jobPath, cleanup := createJobPath()
	defer cleanup()

	s3Client := internal.GetS3Client(&req.S3Config)

	if err := fetchInputs(s3Client, req.Inputs, jobPath); err != nil {
		internal.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err, stdErr := runFFmpeg(req.Commands, jobPath); err != nil {
		internal.WriteConsoleError(w, http.StatusInternalServerError, err.Error(), stdErr)
		return
	}

	if req.Output.InlineContentType != "" {
		streamFirstFile(w, r, jobPath, req.Output.InlineContentType)
		return
	}

	results, err := collectResults(s3Client, req, jobPath)
	if err != nil {
		internal.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	internal.WriteJSON(w, http.StatusOK, api.ProcessResponse{Results: results})
}

func createJobPath() (string, func()) {
	path := filepath.Join(os.TempDir(), uuid.New().String())
	os.MkdirAll(path, 0755)
	return path, func() { os.RemoveAll(path) }
}

func fetchInputs(s3Client *s3.Client, inputs map[string]api.Input, jobPath string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(inputs))

	for name, input := range inputs {
		wg.Add(1)
		go func(name string, input api.Input) {
			defer wg.Done()
			dst := filepath.Join(jobPath, name)
			if err := fetchInput(s3Client, input, dst); err != nil {
				errCh <- fmt.Errorf("input %s: %w", name, err)
			}
		}(name, input)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}
	return nil
}

func fetchInput(s3Client *s3.Client, input api.Input, dst string) error {
	switch {
	case input.S3 != "":
		return internal.DownloadFromS3(s3Client, input.S3, dst)
	case input.HTTP != "":
		return internal.DownloadFromHTTP(input.HTTP, dst)
	case input.Base64 != "":
		return internal.WriteBase64ToFile(input.Base64, dst)
	default:
		return fmt.Errorf("input source missing")
	}
}

func runFFmpeg(commands [][]string, dir string) (error, []string) {
	for _, args := range commands {
		cmd := exec.Command("ffmpeg", sanitizeArgs(args)...)
		cmd.Dir = dir

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return err, strings.Split(strings.TrimRight(stderr.String(), "\n"), "\n")
		}
	}
	return nil, nil
}

func sanitizeArgs(args []string) []string {
	out := []string{"-hide_banner"}
	for _, a := range args {
		if a != "-hide_banner" {
			out = append(out, a)
		}
	}
	return out
}

func collectResults(s3Client *s3.Client, req api.ProcessRequest, jobPath string) (map[string]api.Result, error) {
	results := make(map[string]api.Result)

	files, _ := os.ReadDir(jobPath)
	for _, f := range files {
		if _, isInput := req.Inputs[f.Name()]; isInput {
			continue
		}

		path := filepath.Join(jobPath, f.Name())
		result := api.Result{}

		if req.Output.S3 != "" {
			url, err := internal.UploadToS3(s3Client, req.Output.S3, path, f.Name())
			if err != nil {
				return nil, err
			}
			result.URL = url
		}

		if req.Output.Base64 {
			data, _ := os.ReadFile(path)
			result.Base64 = base64.StdEncoding.EncodeToString(data)
		}

		results[f.Name()] = result
	}
	return results, nil
}

func streamFirstFile(w http.ResponseWriter, r *http.Request, jobPath, contentType string) {
	files, _ := os.ReadDir(jobPath)
	for _, f := range files {
		if !f.IsDir() {
			w.Header().Set("Content-Type", contentType)
			http.ServeFile(w, r, filepath.Join(jobPath, f.Name()))
			return
		}
	}
}
