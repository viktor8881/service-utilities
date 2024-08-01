package tbot

import (
	"encoding/json"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type CustomBot struct {
	*telebot.Bot
}

type Middleware func(next telebot.HandlerFunc) telebot.HandlerFunc

func NewCustomBot(settings telebot.Settings) (*CustomBot, error) {
	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}
	return &CustomBot{Bot: bot}, nil
}

func (b *CustomBot) AddCommandHandler(
	command any,
	in interface{},
	handlerFn func(c telebot.Context, in any) (any, error),
	encodeFn func(c telebot.Context, outDto any) error,
	logger *zap.Logger,
	middlewares ...Middleware,
) {
	wrappedHandler := func(c telebot.Context) error {

		h := &handler{
			command:   command,
			in:        in,
			decodeFn:  decodePayload,
			handlerFn: handlerFn,
			encodeFn:  encodeFn,
			errorFn:   errorFunction,
			logger:    logger,
		}
		return h.ServeTbot(c)

	}

	for _, middleware := range middlewares {
		wrappedHandler = middleware(wrappedHandler)
	}

	b.Handle(command, wrappedHandler)
}

func decodePayload(payload string, in interface{}) error {
	return json.Unmarshal([]byte(payload), in)
}

func EncodeOutput(c telebot.Context, out interface{}) error {
	encoded, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return c.Send(string(encoded))
}

func errorFunction(c telebot.Context, err error, logger *zap.Logger) {
	logger.Error("Error", zap.Error(err))
	if errSend := c.Send("An error occurred"); errSend != nil {
		logger.Error("Error sending response", zap.Error(errSend), zap.Error(err))
	}
}
