/*
slog logger for tests that doesnt log anything,
because we dont want to log anything when testing ¯\_(ツ)_/¯
*/
package slogdiscard

import (
	"context"

	"log/slog"
)

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

type DiscardHandler struct{}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// ignore writing
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// return same handler
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	// return same handler
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// false because writing is ignored
	return false
}
