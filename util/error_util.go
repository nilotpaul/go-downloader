package util

type AppError struct {
	Status int
	Msg    string
	Err    []any
}

func NewAppError(status int, errMsg string, err ...any) *AppError {
	return &AppError{
		Status: status,
		Msg:    errMsg,
		Err:    err,
	}
}

func (e *AppError) Error() string {
	return e.Msg
}
