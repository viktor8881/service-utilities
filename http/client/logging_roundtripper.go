package client

import (
	"bytes"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// LoggingRoundTripper is a http.RoundTripper that logs requests and responses
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

	var requestBody []byte
	if req.Body != nil {
		requestBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	lrt.Logger.Info("httpclient: send request",
		zap.String("url", req.Method+": "+req.URL.String()),
		zap.String("requestBody", string(requestBody)),
	)

	resp, err := lrt.Proxied.RoundTrip(req)
	if err != nil {
		lrt.Logger.Info("httpclient: request error",
			zap.String("url", req.Method+": "+req.URL.String()),
			zap.String("requestBody", string(requestBody)),
			zap.Error(err),
		)

		return nil, err
	}

	if lrt.TurnOnAll {
		duration := time.Since(start)
		lrt.Logger.Info("httpclient: request processed",
			zap.String("url", req.Method+": "+req.URL.String()),
			zap.String("requestBody", string(requestBody)),
			zap.String("StatusResponse", resp.Status),
			zap.Duration("Duration", duration),
		)
	}

	return resp, nil
}
