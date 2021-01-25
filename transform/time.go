package transform

import (
	"time"
)

func NanosecondsToSeconds(ns int64) float64 {
	return float64(ns*int64(time.Nanosecond)) / float64(time.Second)
}

func NanosecondsToMilliseconds(ns int64) int64 {
	return int64(float64(ns*int64(time.Nanosecond)) / float64(time.Millisecond))
}
