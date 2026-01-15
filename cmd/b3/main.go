package main

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
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

func loadSessionKey() (string, error) {
	if value := os.Getenv("WEB_SESSION_KEY"); value != "" {
		return value, nil
	}

	logger.Warn("generating random session key. Please set env WEB_SESSION_KEY")

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

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

	sessionKey, err := loadSessionKey()
	if err != nil {
		logger.Error("failed to generate session key",
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
		Logger:     logger.WithGroup("http"),
		SessionKey: sessionKey,
		Pages:      renderer,
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
