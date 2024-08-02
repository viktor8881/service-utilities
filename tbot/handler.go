package tbot

import (
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
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
	command   any
	in        any
	decodeFn  func(data string, inDto any) error
	handlerFn func(c telebot.Context, in any) (any, error)
	encodeFn  func(c telebot.Context, outDto any) error
	errorFn   func(c telebot.Context, err error, logger *zap.Logger)
	logger    *zap.Logger
}

func (h *handler) ServeTbot(c telebot.Context) error {
	payload := c.Message().Payload

	inDto := reflect.New(reflect.TypeOf(h.in).Elem()).Interface()
	if len(payload) > 0 && h.in != nil && h.decodeFn != nil {
		if err := h.decodeFn(payload, inDto); err != nil {
			h.errorFn(c, err, h.logger)
			return c.Send("Error decoding input")
		}
	}

	outDto, err := h.handlerFn(c, inDto)
	if err != nil {
		h.errorFn(c, err, h.logger)
		return c.Send("Error processing request")
	}

	if h.encodeFn != nil {
		if err := h.encodeFn(c, outDto); err != nil {
			h.errorFn(c, err, h.logger)
			return c.Send("Error encoding output")
		}

		return nil
	}

	return c.Send(outDto)
}
