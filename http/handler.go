package http

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

type ErrorHandlerFunc func(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	logger *zap.Logger,
)

type handler struct {
	path      string
	method    string
	in        any
	decodeFn  func(req *http.Request, inDto any) error
	handlerFn func(ctx context.Context, in any) (any, error)
	encodeFn  func(res http.ResponseWriter, outDto any) error
	errorFn   ErrorHandlerFunc
	logger    *zap.Logger
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != h.method {
		h.errorFn(w, r, &MethodNotAllowedError{}, h.logger)
		return
	}

	inDto := reflect.New(reflect.TypeOf(h.in).Elem()).Interface()
	if h.in != nil {
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
