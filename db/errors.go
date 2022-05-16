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
	return r.FindStringSubmatch(error.Error())[1]
}
