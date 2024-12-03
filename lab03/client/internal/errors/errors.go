package errors

type APIError struct {
	Err        error
	StatusCode int
}

func (a APIError) Error() string {
	return a.Err.Error()
}
