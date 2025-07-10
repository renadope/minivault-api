package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func Serve(server *http.Server, wg *sync.WaitGroup, logger *slog.Logger) error {
	if server == nil {
		return fmt.Errorf("no server specified")
	}
	if wg == nil {
		logger.Warn("No WaitGroup provided")
	}
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		s := <-quit
		logger.Info("shutting down server", "signal", s.String())
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		logger.Info(
			"completing background tasks",
			slog.String("addr", server.Addr),
		)
		if wg != nil {
			wg.Wait()
		}
		shutdownError <- nil

	}()

	logger.Info("starting server",
		slog.String("addr", server.Addr))

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("error shutting down server:%w", err)
	}
	logger.Info("stopping server",
		slog.String("addr", server.Addr))
	return nil

}
