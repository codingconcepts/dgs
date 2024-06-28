package model

import (
	"fmt"
	"strings"
)

type ErrBuilder struct {
	b   strings.Builder
	err error
}

func (b *ErrBuilder) WriteString(format string, args ...any) {
	if b.err != nil {
		return
	}

	_, b.err = b.b.WriteString(fmt.Sprintf(format, args...))
}

func (b *ErrBuilder) String() string {
	return b.b.String()
}

func (b *ErrBuilder) Error() error {
	return b.err
}
