package logger

import (
	"context"
	"log/slog"
	"os"
)

// multi routes:
// INFO and below  -> stdout
// ERROR and above -> stderr
type multi struct {
	infoHandler  slog.Handler
	errorHandler slog.Handler
}

func New() *slog.Logger {
	infoHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	errorHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	return slog.New(&multi{
		infoHandler:  infoHandler,
		errorHandler: errorHandler,
	})
}

func (m *multi) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *multi) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= slog.LevelError {
		return m.errorHandler.Handle(ctx, record)
	}
	return m.infoHandler.Handle(ctx, record)
}

func (m *multi) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &multi{
		infoHandler:  m.infoHandler.WithAttrs(attrs),
		errorHandler: m.errorHandler.WithAttrs(attrs),
	}
}

func (m *multi) WithGroup(name string) slog.Handler {
	return &multi{
		infoHandler:  m.infoHandler.WithGroup(name),
		errorHandler: m.errorHandler.WithGroup(name),
	}
}
