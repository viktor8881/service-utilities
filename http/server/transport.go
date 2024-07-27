package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/form"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

type Transport struct {
	mux *http.ServeMux
}

type Middleware func(http.Handler) http.Handler

func NewTransport(mux *http.ServeMux) *Transport {
	return &Transport{
		mux: mux,
	}
}

func (t *Transport) AddEndpoint(
	path string,
	method string,
	in interface{},
	handlerFn func(ctx context.Context, in interface{}) (interface{}, error),
	logger *zap.Logger,
	middlewares ...Middleware,
) {
	h := &handler{
		path,
		method,
		in,
		t.decodeRequest,
		handlerFn,
		t.encodeResponse,
		ErrorHandler,
		logger,
	}

	wrappedHandler := applyMiddleware(h, middlewares...)

	t.mux.Handle(path, wrappedHandler)
}

func (t *Transport) encodeResponse(res http.ResponseWriter, outDto any) error {
	res.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(res)
	if err := encoder.Encode(outDto); err != nil {
		return err
	}
	return nil
}

func (t *Transport) decodeRequest(r *http.Request, inDto any) error {
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

func applyMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
