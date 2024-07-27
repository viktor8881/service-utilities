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
	mess := fmt.Sprintf("httpCode2user: %d, httpBody2user: %s", e.HttpCode, e.HttpMessage)
	if e.Err != nil {
		mess += "; error: " + e.Err.Error()
	}

	return mess
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

	zapFields := []zap.Field{
		zap.String("url", r.Method+": "+r.URL.String()),
		zap.String("body", bodyStr),
		zap.Int("httpCode2user", code),
		zap.String("httpBody2user", message),
	}
	if err != nil {
		zapFields = append(zapFields, zap.Error(err))
	}

	logger.Error(message, zapFields...)

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
