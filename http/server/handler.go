package server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

type DecodeRequestFunc func(req *http.Request, inDto any) error
type HandlerFunc func(ctx context.Context, in any) (any, error)
type EncodeResponseFunc func(res http.ResponseWriter, outDto any) error
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error, logger *zap.Logger)
type Middleware func(http.Handler) http.Handler

type handler struct {
	path      string
	method    string
	in        any
	decodeFn  DecodeRequestFunc
	handlerFn HandlerFunc
	encodeFn  EncodeResponseFunc
	errorFn   ErrorHandlerFunc
	logger    *zap.Logger
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != h.method {
		err := &CustomError{
			Err:         errors.New("method not allowed"),
			HttpCode:    http.StatusMethodNotAllowed,
			HttpMessage: "method not allowed",
		}

		h.errorFn(w, r, err, h.logger)
		return
	}

	inDto := reflect.New(reflect.TypeOf(h.in).Elem()).Interface()
	if h.in != nil && h.decodeFn != nil {
		if err := h.decodeFn(r, inDto); err != nil {
			h.errorFn(w, r, err, h.logger)
			return
		}
	}

	outDto, err := h.handlerFn(ctx, inDto)
	if err != nil {
		h.errorFn(w, r, err, h.logger)
		return
	}

	if h.encodeFn != nil {
		if err := h.encodeFn(w, outDto); err != nil {
			h.errorFn(w, r, err, h.logger)
			return
		}
	}
}
