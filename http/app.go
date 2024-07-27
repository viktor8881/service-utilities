package http

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	Addr   string
	Mux    *http.ServeMux
	Logger *zap.Logger
}

func NewApp(addr string, logger *zap.Logger) *App {
	return &App{
		Addr:   addr,
		Mux:    http.NewServeMux(),
		Logger: logger,
	}
}

func (a *App) Run() {
	server := &http.Server{
		Addr:    a.Addr,
		Handler: a.Mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Fatal("error when running server:", zap.Error(err))
		}
	}()

	a.Logger.Info("server is run.", zap.String("addr", a.Addr))

	waitForShutdown(server, a.Logger)
}

func waitForShutdown(s *http.Server, logger *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down the server...")
	if err := s.Shutdown(context.Background()); err != nil {
		logger.Fatal("error when shutting down the server:", zap.Error(err))
	}
}
