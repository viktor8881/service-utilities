package tbot

import (
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
	"reflect"
)

type DecodePayloadFunc func(payload string, in interface{}) error
type HandlerFunc func(c telebot.Context, in any) (any, error)
type EncodeResponseFunc func(c telebot.Context, outDto any) error
type ErrorHandlerFunc func(c telebot.Context, err error, logger *zap.Logger)
type Middleware func(next telebot.HandlerFunc) telebot.HandlerFunc

type handler struct {
	command   any
	in        any
	decodeFn  DecodePayloadFunc
	handlerFn HandlerFunc
	encodeFn  EncodeResponseFunc
	errorFn   ErrorHandlerFunc
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
