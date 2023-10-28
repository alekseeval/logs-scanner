package model

const (
	WrongRequestFormat = 5001
	PageNotFound       = 5002
)

type ServerError struct {
	Code        int
	Description string
}

func (e *ServerError) Error() string {
	return e.Description
}
