package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func WriteConsoleError(w http.ResponseWriter, status int, msg string, console []string) {
	WriteJSON(w, status, map[string]any{"error": msg, "console": console})
}

func DownloadFromHTTP(src, dst string) error {
	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("failed to GET %s: %w", src, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}

	return nil
}
