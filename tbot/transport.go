package tbot

import (
	"encoding/json"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Bot struct {
	*telebot.Bot
}

func NewBot(settings telebot.Settings) (*Bot, error) {
	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}
	return &Bot{Bot: bot}, nil
}

func (b *Bot) AddCommandHandler(
	command any,
	in interface{},
	decodeFn DecodePayloadFunc,
	handlerFn HandlerFunc,
	encodeFn EncodeResponseFunc,
	errHandlerFn ErrorHandlerFunc,
	logger *zap.Logger,
	middlewares ...Middleware,
) {
	wrappedHandler := func(c telebot.Context) error {
		h := &handler{
			command:   command,
			in:        in,
			decodeFn:  decodeFn,
			handlerFn: handlerFn,
			encodeFn:  encodeFn,
			errorFn:   errHandlerFn,
			logger:    logger,
		}

		return h.ServeTbot(c)
	}

	for _, middleware := range middlewares {
		wrappedHandler = middleware(wrappedHandler)
	}

	b.Handle(command, wrappedHandler)
}

func DecodePayload(payload string, in interface{}) error {
	return json.Unmarshal([]byte(payload), in)
}

func EncodeResponse(c telebot.Context, out interface{}) error {
	encoded, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return c.Send(string(encoded))
}

func ErrorHandler(c telebot.Context, err error, logger *zap.Logger) {
	logger.Error("tbot: Error", zap.Error(err))
	errSend := c.Send("An error occurred")
	if errSend != nil {
		logger.Error("tbot: Error send error information", zap.Error(errSend))
	}
}
