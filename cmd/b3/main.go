package main

import (
	"log/slog"
	"os"

	"github.com/groundsgg/b3/pkg/log"
)

var logger *slog.Logger

func main() {
	logger = log.BuildRootLogger()

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "hash":
			hashPassword()
			return
		}
	}

	startServer()
}
