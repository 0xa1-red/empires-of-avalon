package model

type CommonResponse struct {
	Status     int
	StatusText string
	Message    string
}

type ErrorResponse struct {
	Status     int
	StatusText string
	Error      error
}

type BuildRequest struct {
	Building string `json:"building"`
	Amount   int    `json:"string,omitempty"`
}
