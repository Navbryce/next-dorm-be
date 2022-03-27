package util

import (
	"time"
)

func ParseTime(val string) (time.Time, error) {
	return time.Parse(time.RFC3339, val)
}
