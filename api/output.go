package api

type Output struct {
	S3                string `json:"s3,omitempty"`
	InlineContentType string `json:"inlineContentType,omitempty"`
	Base64            bool   `json:"base64,omitempty"`
}
