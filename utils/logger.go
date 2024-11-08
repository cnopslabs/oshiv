package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func loggerInit() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	JsonHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	Logger = slog.New(JsonHandler)
}
