package model

const (
	InternalServerError   = 5001
	UnknownDBError        = 5002
	WrongFormatError      = 5003
	NoClusterNameProvided = 5004
	NoNamespaceProvided   = 5005
)

func NewServerErrorByCode(errCode int) *ServerError {
	var sError ServerError
	switch errCode {
	case NoClusterNameProvided:
		sError.Description = "no cluster name provided in request"
	case NoNamespaceProvided:
		sError.Description = "no namespace provided in request"
	default:
		sError.Code = InternalServerError
		sError.Description = "Unexpected error occurred"
	}
	return &sError
}

type ServerError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (e *ServerError) Error() string {
	return e.Description
}
