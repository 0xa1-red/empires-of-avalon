package api

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
