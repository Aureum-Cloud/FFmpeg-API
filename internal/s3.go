package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"ffmpeg/api"
)

func GetS3Client(authentication *api.S3Config) *s3.Client {
	scheme := "http"
	if authentication.UseSSL == nil || *authentication.UseSSL {
		scheme = "https"
	}
	return s3.New(s3.Options{
		Region:       authentication.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(authentication.AccessKey, authentication.SecretKey, ""),
		BaseEndpoint: aws.String(fmt.Sprintf("%s://%s", scheme, authentication.Endpoint)),
		UsePathStyle: true,
	})
}

func DownloadFromS3(client *s3.Client, s3url, dest string) error {
	bucket, key := parseS3(s3url)

	out, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0644)
}

func UploadToS3(client *s3.Client, baseURL, path, name string) (string, error) {
	bucket, prefix := parseS3(baseURL)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	key := strings.TrimSuffix(prefix, "/") + "/" + name

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", err
	}

	return "s3://" + bucket + "/" + key, nil
}

func parseS3(url string) (bucket, key string) {
	trimmed := strings.TrimPrefix(url, "s3://")
	parts := strings.SplitN(trimmed, "/", 2)
	return parts[0], parts[1]
}
