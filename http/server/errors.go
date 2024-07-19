package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type CustomError struct {
	HttpCode    int
	HttpMessage string
	Err         error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("httpCode2user: %d; httpBody2user: %s; error: %s", e.HttpMessage, e.Err.Error())
}

func ErrorHandler(w http.ResponseWriter,
	r *http.Request,
	err error,
	logger *zap.Logger,
) {
	var customError *CustomError

	var code int
	var message string

	switch {
	case errors.As(err, &customError):
		code = customError.HttpCode
		message = customError.HttpMessage
	default:
		code = http.StatusInternalServerError
		message = "internal server error"
	}

	var bodyStr string
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		bodyStr = string(body)
	} else {
		bodyStr = r.URL.Query().Encode()
	}

	logger.Error(message,
		zap.String("url", r.Method+": "+r.URL.String()),
		zap.String("body", bodyStr),
		zap.Int("httpCode2user", code),
		zap.String("httpBody2user", message),
		zap.Error(err))

	w.WriteHeader(code)

	if code >= 400 && code < 500 {
		errorResponse := struct {
			Message string `json:"message"`
		}{
			Message: message,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&errorResponse)
	}

}
