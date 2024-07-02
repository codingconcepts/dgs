package test

import "github.com/rs/zerolog"

type NilWriter struct{}

func (w NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func NewNilLogger() zerolog.Logger {
	return zerolog.New(NilWriter{})
}
