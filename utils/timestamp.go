package utils

import (
	"time"
)

func NanosecondsToSeconds(timestamp int64) float64 {
	return float64(timestamp) / float64(time.Second)
}
