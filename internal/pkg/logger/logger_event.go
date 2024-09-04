package logger

import (
	"time"

	"github.com/gookit/slog"
)

// EventFn type
type EventFn func(e EventArgs) error

// event struct
type event struct {
	level   slog.Level
	eventFn EventFn
}

// EventArgs struct
type EventArgs struct {
	Time                    time.Time
	Args                    []any
	Level, Channel, Message string
}

// Close func
func (e event) Close() error {
	return nil
}

// Flush func
func (e event) Flush() error {
	return nil
}

// IsHandling func
func (e event) IsHandling(level slog.Level) bool {
	return level <= e.level
}

// Handle func
func (e event) Handle(r *slog.Record) error {
	args := EventArgs{Time: r.Time, Args: r.Args, Level: r.Level.Name(), Channel: r.Channel, Message: r.Message}
	return e.eventFn(args)
}
