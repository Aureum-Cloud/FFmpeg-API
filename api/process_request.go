package api

type ProcessRequest struct {
	S3Config S3Config         `json:"s3Config"`
	Inputs   map[string]Input `json:"input"`
	Commands [][]string       `json:"commands,omitempty"`
	Output   Output           `json:"output,omitempty"`
}
