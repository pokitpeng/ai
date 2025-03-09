package main

import (
	"log/slog"
	"os"

	"github.com/pokitpeng/ai/cmd/ai"
)

func main() {
	// if set env LOG_LEVEL=debug, then set level to debug
	var level slog.Level
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Execute AI command line tool
	ai.Execute()
}
