package random

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func Int(min, max string) (int64, error) {
	parsedMin, err := strconv.ParseInt(min, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing min: %w", err)
	}

	parsedMax, err := strconv.ParseInt(max, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing max: %w", err)
	}

	return rand.Int63n(parsedMax-parsedMin) + parsedMin, nil
}

func Float(min, max string) (float64, error) {
	parsedMin, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing min: %w", err)
	}

	parsedMax, err := strconv.ParseFloat(max, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing max: %w", err)
	}

	return parsedMin + rand.Float64()*(parsedMax-parsedMin), nil
}

func Timestamp(min, max string) (time.Time, error) {
	parsedMin, err := time.Parse(time.RFC3339, min)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing min: %w", err)
	}

	parsedMax, err := time.Parse(time.RFC3339, max)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing max: %w", err)
	}

	minUnix := parsedMin.Unix()
	maxUnix := parsedMax.Unix()
	delta := maxUnix - minUnix

	randUnix := minUnix + rand.Int63n(delta)
	return time.Unix(randUnix, 0), nil

}

func UUID() string {
	return uuid.NewString()
}