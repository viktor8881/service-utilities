package tbot

import (
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"

	"time"
)

func LoggerMiddleware(logger *zap.Logger) func(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			start := time.Now()

			logger.Info("tbot: incoming request",
				zap.String("user", c.Sender().FirstName+" "+c.Sender().LastName),
				zap.String("text", c.Text()),
			)

			err := next(c)

			duration := time.Since(start)
			logger.Info("tbot: request processed",
				zap.Int64("user_id", c.Sender().ID),
				zap.String("user", c.Sender().FirstName+" "+c.Sender().LastName),
				zap.String("text", c.Text()),
				zap.Duration("duration", duration),
				zap.Error(err),
			)

			return err
		}
	}
}
