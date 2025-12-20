package api

type S3Config struct {
	UseSSL    *bool  `json:"useSSL,omitempty"`
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}
