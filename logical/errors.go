package logical

func NewCodedError(status int, msg string) CodedError {
	return CodedError{
		Status:  status,
		Message: msg,
	}
}

type CodedError struct {
	Status  int
	Message string
}

func (e CodedError) Error() string {
	return e.Message
}

func (e CodedError) Code() int {
	return e.Status
}