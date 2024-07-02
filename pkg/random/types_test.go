package random

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	cases := []struct {
		name    string
		min     string
		max     string
		expFunc func(int64) bool
		expErr  error
	}{
		{
			name: "equal min max",
			min:  "10",
			max:  "10",
			expFunc: func(i int64) bool {
				return i == 10
			},
		},
		{
			name: "different min max",
			min:  "10",
			max:  "100",
			expFunc: func(i int64) bool {
				return i >= 10 && i < 100
			},
		},
		{
			name:   "invalid min",
			min:    "a",
			max:    "100",
			expErr: &strconv.NumError{Num: "a", Func: "ParseInt", Err: errors.New("invalid syntax")},
		},
		{
			name:   "invalid max",
			min:    "10",
			max:    "a",
			expErr: &strconv.NumError{Num: "a", Func: "ParseInt", Err: errors.New("invalid syntax")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := Int(c.min, c.max)
			if c.expErr != nil {
				assert.Equal(t, c.expErr, errors.Unwrap(err))
				return
			}

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestFloat(t *testing.T) {
	cases := []struct {
		name    string
		min     string
		max     string
		expFunc func(f float64) bool
		expErr  error
	}{
		{
			name: "equal min max",
			min:  "10",
			max:  "10",
			expFunc: func(f float64) bool {
				return f == 10
			},
		},
		{
			name: "different min max",
			min:  "10",
			max:  "100",
			expFunc: func(f float64) bool {
				return f >= 10 && f < 100
			},
		},
		{
			name:   "invalid min",
			min:    "a",
			max:    "100",
			expErr: &strconv.NumError{Num: "a", Func: "ParseFloat", Err: errors.New("invalid syntax")},
		},
		{
			name:   "invalid max",
			min:    "10",
			max:    "a",
			expErr: &strconv.NumError{Num: "a", Func: "ParseFloat", Err: errors.New("invalid syntax")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := Float(c.min, c.max)
			if c.expErr != nil {
				assert.Equal(t, c.expErr, errors.Unwrap(err))
				return
			}

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestTimestamp(t *testing.T) {
	cases := []struct {
		name    string
		min     string
		max     string
		expFunc func(time.Time) bool
		expErr  error
	}{
		{
			name: "equal min max",
			min:  "2024-01-01T01:01:01Z",
			max:  "2024-01-01T01:01:01Z",
			expFunc: func(t time.Time) bool {
				return t.Equal(time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC))
			},
		},
		{
			name: "different min max",
			min:  "2024-01-01T01:01:01Z",
			max:  "2025-01-01T01:01:01Z",
			expFunc: func(t time.Time) bool {
				min := time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC)
				max := time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC)

				return (t.After(min) || t.Equal(min)) && t.Before(max)
			},
		},
		{
			name:   "invalid min",
			min:    "a",
			max:    "2025-01-01T01:01:01Z",
			expErr: &time.ParseError{Layout: "2006-01-02T15:04:05Z07:00", Value: "a", LayoutElem: "2006", ValueElem: "a"},
		},
		{
			name:   "invalid max",
			min:    "2024-01-01T01:01:01Z",
			max:    "a",
			expErr: &time.ParseError{Layout: "2006-01-02T15:04:05Z07:00", Value: "a", LayoutElem: "2006", ValueElem: "a"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := Timestamp(c.min, c.max)
			if c.expErr != nil {
				assert.Equal(t, c.expErr, errors.Unwrap(err))
				return
			}

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestBytes(t *testing.T) {
	cases := []struct {
		name    string
		min     string
		max     string
		expFunc func(b []byte) bool
		expErr  error
	}{
		{
			name: "equal min max",
			min:  "10",
			max:  "10",
			expFunc: func(b []byte) bool {
				return len(b) == 10
			},
		},
		{
			name: "different min max",
			min:  "10",
			max:  "100",
			expFunc: func(b []byte) bool {
				return len(b) >= 10 && len(b) < 100
			},
		},
		{
			name:   "invalid min",
			min:    "a",
			max:    "100",
			expErr: &strconv.NumError{Num: "a", Func: "ParseInt", Err: errors.New("invalid syntax")},
		},
		{
			name:   "invalid max",
			min:    "10",
			max:    "a",
			expErr: &strconv.NumError{Num: "a", Func: "ParseInt", Err: errors.New("invalid syntax")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := Bytes(c.min, c.max)
			if c.expErr != nil {
				assert.Equal(t, c.expErr, errors.Unwrap(err))
				return
			}

			assert.True(t, c.expFunc(act))
		})
	}
}
