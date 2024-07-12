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
	fErrorHandler func(w http.ResponseWriter, r *http.Request, err error, logger *zap.Logger),
) {
	h := &handler{
		path,
		method,
		in,
		t.decodeRequest,
		handlerFn,
		t.encodeResponse,
		fErrorHandler,
		logger,
	}

	t.mux.Handle(path, h)
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
		return &DecodeEncodeError{
			Err:       errors.New("input must be a pointer"),
			Mess2user: "unable to decode request",
			Code2user: http.StatusInternalServerError,
		}
	}

	if r.Method == http.MethodGet {
		decoder := form.NewDecoder()
		if err := decoder.Decode(inDto, r.URL.Query()); err != nil {
			return &DecodeEncodeError{
				Err:       err,
				Mess2user: "unable to decode request",
				Code2user: http.StatusBadRequest,
			}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(inDto); err != nil {
			return &DecodeEncodeError{
				Err:       err,
				Mess2user: "unable to decode request",
				Code2user: http.StatusInternalServerError,
			}
		}
	}

	return nil
}

//
//func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) (int, string) {
//	var decodeEncodeError *DecodeEncodeError
//	var methodNotAllowedError *MethodNotAllowedError
//
//	var code int
//	var message string
//
//	switch {
//	case errors.Is(err, context.Canceled):
//		code = http.StatusRequestTimeout
//		message = "request canceled"
//	case errors.Is(err, context.DeadlineExceeded):
//		code = http.StatusGatewayTimeout
//		message = "request deadline exceeded"
//	case errors.As(err, &methodNotAllowedError):
//		code = http.StatusMethodNotAllowed
//		message = "method not allowed"
//	case errors.As(err, &decodeEncodeError):
//		code = decodeEncodeError.Code2user
//		message = decodeEncodeError.Mess2user
//	}
//
//	return code, message
//}
