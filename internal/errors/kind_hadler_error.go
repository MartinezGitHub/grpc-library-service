package errors

import "strconv"

type KindHandlerError struct {
	message string
	status  int
}

func (e *KindHandlerError) Error() string {
	return e.message + " status_code: " + strconv.Itoa(e.status)
}

func (e *KindHandlerError) Status() int {
	return e.status
}

func NewKindHandlerError(message string, status int) *KindHandlerError {
	return &KindHandlerError{
		message: message,
		status:  status,
	}
}
