package logs

import (
	"bytes"
	"context"
	"log/slog"
)

type SlogWriter struct {
	Logger *slog.Logger
	Level  slog.Level
	App    string
}

func (sw *SlogWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		sw.Logger.Log(
			context.Background(),
			sw.Level,
			string(bytes.TrimSpace(p)),
			slog.String("APP", sw.App),
		)
	}
	return len(p), nil
}
