package main

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/groundsgg/b3"
	"github.com/groundsgg/b3/internal/web"
	"github.com/groundsgg/b3/pkg/log"
)

var logger *slog.Logger

func main() {
	// init logger
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	logger = log.BuildRootLogger(&slog.HandlerOptions{
		Level: logLevel,
	})

	// load templates
	renderer, err := b3.NewRenderer()
	if err != nil {
		logger.Error("failed to load html templates",
			"err", err,
		)
		return
	}

	// create server
	server := web.NewServer(web.ServerConfig{
		ListenAddr: cmp.Or(
			os.Getenv("WEB_LISTEN_ADDR"),
			":8080",
		),
		BaseURL: cmp.Or(
			os.Getenv("WEB_BASE_URL"),
			"http://localhost:8080",
		),
		Logger: logger.WithGroup("http"),
		Pages:  renderer,
	})

	// create shutdown context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	go func() {
		<-ctx.Done()
		ctx2, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := server.Stop(ctx2); err != nil {
			logger.Error("web server shutdown error",
				"err", err,
			)
		}
	}()

	// start web server
	if err := server.Start(); err != nil {
		logger.Error("failed to start the web server",
			"err", err,
		)
	}
}
