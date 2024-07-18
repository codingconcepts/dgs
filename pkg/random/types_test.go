package random

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	cases := []struct {
		name    string
		min     int64
		max     int64
		expFunc func(int64) bool
	}{
		{
			name: "equal min max",
			min:  10,
			max:  10,
			expFunc: func(i int64) bool {
				return i == 10
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(i int64) bool {
				return i >= 10 && i < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(i int64) bool {
				return i >= 10 && i < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Int(c.min, c.max)
			assert.True(t, c.expFunc(act))
		})
	}
}

func TestFloat(t *testing.T) {
	cases := []struct {
		name    string
		min     float64
		max     float64
		expFunc func(f float64) bool
	}{
		{
			name: "equal min max",
			min:  10,
			max:  10,
			expFunc: func(f float64) bool {
				return f == 10
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(f float64) bool {
				return f >= 10 && f < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(f float64) bool {
				return f >= 10 && f < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Float(c.min, c.max)
			assert.True(t, c.expFunc(act))
		})
	}
}

func TestTimestamp(t *testing.T) {
	cases := []struct {
		name    string
		min     time.Time
		max     time.Time
		expFunc func(time.Time) bool
	}{
		{
			name: "equal min max",
			min:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			max:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			expFunc: func(t time.Time) bool {
				return t.Equal(time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC))
			},
		},
		{
			name: "different min max",
			min:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			max:  time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC),
			expFunc: func(t time.Time) bool {
				min := time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC)
				max := time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC)

				return (t.After(min) || t.Equal(min)) && t.Before(max)
			},
		},
		{
			name: "min gt max",
			min:  time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC),
			max:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			expFunc: func(t time.Time) bool {
				min := time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC)
				max := time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC)

				return (t.After(min) || t.Equal(min)) && t.Before(max)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Timestamp(c.min, c.max)
			assert.True(t, c.expFunc(act))
		})
	}
}

func TestBytes(t *testing.T) {
	cases := []struct {
		name    string
		min     int64
		max     int64
		expFunc func(b []byte) bool
	}{
		{
			name: "equal min max",
			min:  10,
			max:  10,
			expFunc: func(b []byte) bool {
				return len(b) == 10
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(b []byte) bool {
				return len(b) >= 10 && len(b) < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(b []byte) bool {
				return len(b) >= 10 && len(b) < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := Bytes(c.min, c.max)
			assert.NoError(t, err)

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		name    string
		min     int64
		max     int64
		expFunc func(s string) bool
	}{
		{
			name: "equal min max",
			min:  10,
			max:  10,
			expFunc: func(s string) bool {
				return len(s) == 10
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(s string) bool {
				return len(s) >= 10 && len(s) < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(s string) bool {
				return len(s) >= 10 && len(s) < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := String(c.min, c.max)

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestBitString(t *testing.T) {
	cases := []struct {
		name    string
		min     int64
		max     int64
		expFunc func(b []byte) bool
	}{
		{
			name: "equal min max",
			min:  10,
			max:  10,
			expFunc: func(b []byte) bool {
				return len(b) == 10
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(b []byte) bool {
				return len(b) >= 10 && len(b) < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(b []byte) bool {
				return len(b) >= 10 && len(b) < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := BitString(c.min, c.max)

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestInterval(t *testing.T) {
	cases := []struct {
		name    string
		min     time.Duration
		max     time.Duration
		expFunc func(d time.Duration) bool
	}{
		{
			name: "equal min max",
			min:  time.Hour,
			max:  time.Hour,
			expFunc: func(d time.Duration) bool {
				return d == time.Hour
			},
		},
		{
			name: "different min max",
			min:  time.Minute,
			max:  time.Hour,
			expFunc: func(d time.Duration) bool {
				return d >= time.Minute && d < time.Hour
			},
		},
		{
			name: "min gt max",
			min:  time.Hour,
			max:  time.Minute,
			expFunc: func(d time.Duration) bool {
				return d >= time.Minute && d < time.Hour
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Interval(c.min, c.max)

			assert.True(t, c.expFunc(act))
		})
	}
}

func TestArray(t *testing.T) {
	cases := []struct {
		name    string
		min     int64
		max     int64
		expFunc func(a []any) bool
	}{
		{
			name: "equal min max",
			min:  3,
			max:  3,
			expFunc: func(a []any) bool {
				return len(a) == 3
			},
		},
		{
			name: "different min max",
			min:  10,
			max:  100,
			expFunc: func(a []any) bool {
				return len(a) >= 10 && len(a) < 100
			},
		},
		{
			name: "min gt max",
			min:  100,
			max:  10,
			expFunc: func(a []any) bool {
				return len(a) >= 10 && len(a) < 100
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Array(c.min, c.max, "${fruit}")

			assert.True(t, c.expFunc(act))
		})
	}
}
