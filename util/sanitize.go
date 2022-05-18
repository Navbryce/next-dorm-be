package util

import (
	"github.com/microcosm-cc/bluemonday"
	"html"
)

var XSSPolicy = bluemonday.UGCPolicy()

// XSSSanitize sanitizes of HTML and returns the unescaped HTML
func XSSSanitize(val string) string {
	return html.UnescapeString(XSSPolicy.Sanitize(val))
}
