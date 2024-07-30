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
	handlerFn func(c telebot.Context, in any) (any, error),
	logger *zap.Logger,
	middlewares ...Middleware,
) {
	wrappedHandler := func(c telebot.Context) error {

		h := &handler{
			command:   command,
			decodeFn:  decodePayload,
			handlerFn: handlerFn,
			//encodeFn: decodePayload,
			errorFn: errorFunction,
			logger:  logger,
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

// Пример функции encodeFn
func encodeOutput(out interface{}) (string, error) {
	encoded, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// Пример функции handlerFn
//func handlerFunction(ctx telebot.Context, in interface{}) (interface{}, error) {
//	req := in.(*ListUserByEmailRequest)
//	if req.Email == "someemail@dom.com" {
//		return "User found", nil
//	}
//	return "User not found", nil
//}

func errorFunction(c telebot.Context, err error, logger *zap.Logger) {
	logger.Error("Error", zap.Error(err))
	c.Send("An error occurred")
}
