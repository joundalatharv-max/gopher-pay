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
	// also write error-level logs to a file
	errorFileHandler slog.Handler
}

func New() *slog.Logger {
	infoHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	errorHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	// Try to open stderr.log for appending; if it fails, continue without file handler
	var errorFileHandler slog.Handler
	if f, err := os.OpenFile("stderr.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		errorFileHandler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelError})
	}

	return slog.New(&multi{
		infoHandler:      infoHandler,
		errorHandler:     errorHandler,
		errorFileHandler: errorFileHandler,
	})
}

func (m *multi) Enabled(ctx context.Context, level slog.Level) bool {
	// Only enable if either underlying handler would handle this level
	if m.infoHandler != nil && m.infoHandler.Enabled(ctx, level) {
		return true
	}
	if m.errorHandler != nil && m.errorHandler.Enabled(ctx, level) {
		return true
	}
	return false
}

func (m *multi) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= slog.LevelError {
		// send to stderr
		if err := m.errorHandler.Handle(ctx, record); err != nil {
			return err
		}
		// also send to file (if configured)
		if m.errorFileHandler != nil {
			return m.errorFileHandler.Handle(ctx, record)
		}
		return nil
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
