package db

import "strings"

func IsDupKeyErr(error error) bool {
	return strings.Contains(error.Error(), "Duplicate")
}
