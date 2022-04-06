package util

import (
	"strconv"
	"time"
)

func ParseTime(val string) (time.Time, error) {
	return time.Parse(time.RFC3339, val)
}

func ParseId(val string) (int64, *HTTPError) {
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, MalformedIdHTTPErr
	}
	return id, nil
}
