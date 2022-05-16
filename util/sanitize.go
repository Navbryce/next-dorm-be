package util

import (
	"github.com/microcosm-cc/bluemonday"
)

var XSSPolicy = bluemonday.UGCPolicy()

func XSSSanitize(val string) string {
	return XSSPolicy.Sanitize(val)
}
