package logger

import (
	"os"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
)

// Logger struct
type Logger struct {
}

// New logger
func New(file string) *Logger {
	option := []handler.ConfigFn{handler.WithBuffSize(0)}
	h, err := handler.NewTimeRotateFile(file, rotatefile.EveryDay, option...)
	if err == nil {
		slog.Configure(func(logger *slog.SugaredLogger) {
			hostname, _ := os.Hostname()
			logger.BackupArgs = true
			logger.ChannelName = hostname
		})

		slog.AddHandler(h)
	}

	return &Logger{}
}

// WithNotify func
func (l *Logger) WithNotify(level slog.Level, fn EventFn) {
	slog.PushHandler(event{level: level, eventFn: fn})
}
