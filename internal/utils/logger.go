package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// Note: Logger will most likely need to initialize first, if other utils.inits
// are created, will need to handle initialization order elsewhere
func init() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	JsonHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	Logger = slog.New(JsonHandler)
}
