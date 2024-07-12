package server

type DecodeEncodeError struct {
	Err       error
	Mess2user string
	Code2user int
}

func (e *DecodeEncodeError) Error() string {
	return e.Err.Error()
}

type MethodNotAllowedError struct{ error }

func (e *MethodNotAllowedError) Error() string {
	return "method not allowed"
}
