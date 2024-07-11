package random

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	earthRadiusMiles = 3958.8
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

	if parsedMin == parsedMax {
		return parsedMin, nil
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

	if parsedMin.Equal(parsedMax) {
		return parsedMin, nil
	}

	minUnix := parsedMin.Unix()
	maxUnix := parsedMax.Unix()
	delta := maxUnix - minUnix

	randUnix := minUnix + rand.Int63n(delta)
	return time.Unix(randUnix, 0), nil
}

func Point(lat, lon, radiusMiles float64) (float64, float64, error) {
	randomDistance := (rand.Float64() * radiusMiles) / earthRadiusMiles
	randomBearing := rand.Float64() * 2 * math.Pi

	latRad := degreesToRadians(lat)
	lonRad := degreesToRadians(lon)

	sinLatRad := math.Sin(latRad)
	cosLatRad := math.Cos(latRad)
	sinRandomDistance := math.Sin(randomDistance)
	cosRandomDistance := math.Cos(randomDistance)
	cosRandomBearing := math.Cos(randomBearing)
	sinRandomBearing := math.Sin(randomBearing)

	newLatRad := math.Asin(sinLatRad*cosRandomDistance + cosLatRad*sinRandomDistance*cosRandomBearing)

	newLonRad := lonRad + math.Atan2(
		sinRandomBearing*sinRandomDistance*cosLatRad,
		cosRandomDistance-sinLatRad*math.Sin(newLatRad),
	)

	return radiansToDegrees(newLatRad), radiansToDegrees(newLonRad), nil
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func Bytes(min, max string) ([]byte, error) {
	n, err := Int(min, max)
	if err != nil {
		return nil, err
	}

	result := make([]byte, n)

	_, err = crand.Read(result)
	if err != nil {
		return nil, fmt.Errorf("creating random bytes: %w", err)
	}

	return result, nil
}

func UUID() string {
	return uuid.NewString()
}
