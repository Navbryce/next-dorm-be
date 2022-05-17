package dao

import "database/sql"

type NullInt64 struct {
	sql.NullInt64
}

// AsInt if parent is nil, returns -1
func (ni *NullInt64) AsInt() int64 {
	if !ni.NullInt64.Valid {
		return -1
	}
	return ni.NullInt64.Int64
}
