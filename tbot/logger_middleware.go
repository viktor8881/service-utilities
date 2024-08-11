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

			sender := c.Sender()
			var senderName = "unknown"
			var senderUserID int64 = 0
			if sender != nil {
				senderName = sender.FirstName + " " + sender.LastName
				senderUserID = sender.ID
			}

			logger.Info("tbot: incoming request",
				zap.String("user", senderName),
				zap.String("text", c.Text()),
			)

			err := next(c)

			duration := time.Since(start)
			logger.Info("tbot: request processed",
				zap.Int64("user_id", senderUserID),
				zap.String("user", senderName),
				zap.String("text", c.Text()),
				zap.Duration("duration", duration),
				zap.Error(err),
			)

			return err
		}
	}
}
