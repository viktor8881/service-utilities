package server

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/form"
	"go.uber.org/zap"
	"io"
	"net/http"
	"reflect"
)

type Transport struct {
	mux *http.ServeMux
}

func NewTransport(mux *http.ServeMux) *Transport {
	return &Transport{
		mux: mux,
	}
}

func (t *Transport) AddEndpoint(
	path string,
	method string,
	in interface{},
	decReqFn DecodeRequestFunc,
	handlerFn HandlerFunc,
	encResFn EncodeResponseFunc,
	errHandlerFn ErrorHandlerFunc,
	logger *zap.Logger,
	middlewares ...Middleware,
) {
	h := &handler{
		path,
		method,
		in,
		decReqFn,
		handlerFn,
		encResFn,
		errHandlerFn,
		logger,
	}

	wrappedHandler := applyMiddleware(h, middlewares...)

	t.mux.Handle(path, wrappedHandler)
}

func EncodeResponse(res http.ResponseWriter, outDto any) error {
	res.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(res)
	if err := encoder.Encode(outDto); err != nil {
		return err
	}
	return nil
}

func DecodeRequest(r *http.Request, inDto any) error {
	if reflect.TypeOf(inDto).Kind() != reflect.Ptr {
		return &CustomError{
			Err:         errors.New("input must be a pointer"),
			HttpCode:    http.StatusInternalServerError,
			HttpMessage: "unable to decode request",
		}
	}

	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		decoder := form.NewDecoder()
		if err := decoder.Decode(inDto, r.URL.Query()); err != nil {
			return &CustomError{
				Err:         err,
				HttpMessage: "unable to decode request",
				HttpCode:    http.StatusBadRequest,
			}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(inDto); err != nil {
			return &CustomError{
				Err:         err,
				HttpMessage: "unable to decode request",
				HttpCode:    http.StatusInternalServerError,
			}
		}
	}

	return nil
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
		defer func() {
			err := r.Body.Close()
			if err != nil {
				logger.Error("httpserver: error while closing request body", zap.Error(err))
			}
		}()

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

	logger.Error("httpserver: error "+message, zapFields...)

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

func applyMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
