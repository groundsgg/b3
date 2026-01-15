// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"log/slog"
	"os"

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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hash":
			hashPassword()
			return
		}
	}

	startServer()
}
