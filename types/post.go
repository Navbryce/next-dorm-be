package types

import "time"

type Community struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type Visibility string

const (
	VisibilityNormal Visibility = "NORMAL"
	VisibilityHidden            = "HIDDEN"
)

type Post struct {
	Id          int64
	Creator     *DisplayableUser
	Content     string
	NumVotes    int
	VoteTotal   int
	Communities []*Community
	Visibility  Visibility
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
