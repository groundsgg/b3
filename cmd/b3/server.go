package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/groundsgg/b3"
	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/config"
	"github.com/groundsgg/b3/internal/web"
)

func createServer() *web.Server {
	err := config.Load()
	if err != nil {
		logger.Error("failed to load config envs",
			"err", err,
		)
		return nil
	}

	renderer, err := b3.NewRenderer()
	if err != nil {
		logger.Error("failed to load html templates",
			"err", err,
		)
		return nil
	}

	ah, err := auth.GetAuthHandler()
	if err != nil {
		logger.Error("failed to load auth handler",
			"err", err,
		)
		return nil
	}
	logger.Info("using auth handler", "type", ah.Type())

	server, err := web.NewServer(web.ServerConfig{
		ListenAddr:  config.GetConfig().Web.ListenAddr,
		BaseURL:     config.GetConfig().Web.BaseURL,
		Logger:      logger.WithGroup("http"),
		SessionKey:  []byte(config.GetConfig().Web.SessionKey),
		Pages:       renderer,
		AuthHandler: ah,
	})
	if err != nil {
		logger.Error("failed to create server",
			"err", err,
		)
		return nil
	}
	return server
}

func startServer() {
	logger.Info("starting server application")

	server := createServer()
	if server == nil {
		return
	}

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

	if err := server.Start(); err != nil {
		logger.Error("failed to start the web server",
			"err", err,
		)
	}
}
