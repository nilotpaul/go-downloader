package util

type AppError struct {
	Status int
	Msg    string
	Err    []interface{}
}

func NewAppError(status int, errMsg string, err ...interface{}) *AppError {
	return &AppError{
		Status: status,
		Msg:    errMsg,
		Err:    err,
	}
}

func (e *AppError) Error() string {
	return e.Msg
}
