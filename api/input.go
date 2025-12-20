package api

type Input struct {
	S3     string `json:"s3,omitempty"`
	HTTP   string `json:"http,omitempty"`
	Base64 string `json:"base64,omitempty"`
}
