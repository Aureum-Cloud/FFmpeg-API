package api

type ProcessResponse struct {
	Results map[string]Result `json:"results"`
}
