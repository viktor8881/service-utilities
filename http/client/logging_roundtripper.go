package simplehttp

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

// LoggingRoundTripper is an http.RoundTripper that logs requests and responses
type LoggingRoundTripper struct {
	Proxied   http.RoundTripper
	Logger    *zap.Logger
	TurnOnAll bool
}

func NewLoggingRoundTripper(proxied http.RoundTripper, logger *zap.Logger, turnOnAll bool) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		Proxied:   proxied,
		Logger:    logger,
		TurnOnAll: turnOnAll,
	}
}

func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := lrt.Proxied.RoundTrip(req)
	if err != nil {
		lrt.Logger.Info("inner request error",
			zap.String("Method", req.Method),
			zap.String("URL", req.URL.String()),
			zap.Error(err),
		)
		return nil, err
	}

	if lrt.TurnOnAll {
		duration := time.Since(start)
		lrt.Logger.Info("inner request result",
			zap.String("Method", req.Method),
			zap.String("URL", req.URL.String()),
			zap.String("Status", resp.Status),
			zap.Duration("Duration", duration),
		)
	}

	return resp, nil
}
