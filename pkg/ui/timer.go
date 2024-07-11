package ui

import (
	"time"

	"github.com/rs/zerolog"
)

type Timer struct {
	name   string
	start  time.Time
	logger zerolog.Logger
}

func NewTimer(name string, logger zerolog.Logger) *Timer {
	return &Timer{
		name:   name,
		start:  time.Now(),
		logger: logger,
	}
}

func (t *Timer) Log() {
	elapsed := time.Since(t.start)
	t.logger.Debug().Str("duration", elapsed.String()).Msg(t.name)
}
