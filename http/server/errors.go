package server

import (
	"fmt"
)

type CustomError struct {
	HttpCode    int
	HttpMessage string
	Err         error
}

func (e *CustomError) Error() string {
	mess := fmt.Sprintf("httpCode2user: %d, httpBody2user: %s", e.HttpCode, e.HttpMessage)
	if e.Err != nil {
		mess += "; error: " + e.Err.Error()
	}

	return mess
}
