package random

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

const (
	earthRadiusMiles = 3958.8
)

var (
	ascii    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	asciiLen = len(ascii) - 1
)

func BitString(min, max int64) []byte {
	size := int(Int(min, max))
	result := make([]byte, size)

	for i := 0; i < size; i++ {
		n := Int(0, 10)
		if n%2 == 0 {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}

	return result
}

func Int(min, max int64) int64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return rand.Int63n(max-min) + min
}

func Float(min, max float64) float64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Float64()*(max-min)
}

func Timestamp(min, max time.Time) time.Time {
	if min.Equal(max) {
		return min
	}

	if min.After(max) {
		min, max = max, min
	}

	minUnix := min.Unix()
	maxUnix := max.Unix()
	delta := maxUnix - minUnix

	randUnix := minUnix + rand.Int63n(delta)
	return time.Unix(randUnix, 0)
}

func Interval(min, max time.Duration) time.Duration {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	diff := max - min
	randomDiff := time.Duration(rand.Int63n(int64(diff)))

	return min + randomDiff
}

func Point(lat, lon, radiusMiles float64) (float64, float64) {
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

	return radiansToDegrees(newLatRad), radiansToDegrees(newLonRad)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func Bytes(min, max int64) ([]byte, error) {
	n := Int(min, max)
	result := make([]byte, n)

	_, err := crand.Read(result)
	if err != nil {
		return nil, fmt.Errorf("creating random bytes: %w", err)
	}

	return result, nil
}

func String(min, max int64) string {
	size := Int(min, max)
	result := make([]rune, size)

	for i := 0; i < int(size); i++ {
		result[i] = rune(ascii[rand.Intn(asciiLen)])
	}

	return string(result)
}

func Array(min, max int64, value string) []any {
	size := Int(min, max)

	result := make([]any, size)
	for i := 0; i < int(size); i++ {
		v, ok := Replacements[value]
		result[i] = lo.Ternary(ok, v(), nil)
	}

	return result
}

func UUID() string {
	return uuid.NewString()
}
