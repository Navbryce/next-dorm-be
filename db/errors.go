package db

import (
	"github.com/go-sql-driver/mysql"
	"regexp"
	"strings"
)

func IsDupKeyErr(error *mysql.MySQLError) bool {
	return strings.Contains(error.Error(), "Duplicate")
}
func GetDupKey(error *mysql.MySQLError) string {
	r := regexp.MustCompile(`(for key ')((.)+)(')`)
	match := r.FindString(error.Error())[9:]
	return match[7 : len(match)-1]
}
