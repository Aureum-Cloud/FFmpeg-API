package internal

import (
	"encoding/base64"
	"os"
)

func WriteBase64ToFile(data, path string) error {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}
