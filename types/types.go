package types

import "time"

type Community struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type Visibility string

const (
	NORMAL Visibility = "NORMAL"
	HIDDEN            = "HIDDEN"
)

type Post struct {
	Id          int64
	Content     string
	Communities []*Community
	Visibility  Visibility
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
}
