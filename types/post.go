package types

import "time"

type Community struct {
	Id        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ParentId  int64  `db:"parent_id" json:"parentId"`
	HasParent bool   `db:"has_parent" json:"hasParent"` // TODO: Add has parent
}

type Visibility string

const (
	VisibilityNormal Visibility = "NORMAL"
	VisibilityHidden            = "HIDDEN"
)

type ContentMetadata struct {
	Id         int64            `json:"-"`
	Creator    *DisplayableUser `json:"creator"`
	Visibility Visibility       `json:"visibility"`
	NumVotes   int              `json:"numVotes"`
	VoteTotal  int              `json:"voteTotal"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

type Post struct {
	*ContentMetadata
	Id          int64        `json:"id"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Communities []*Community `json:"communities"`
}

type Comment struct {
	*ContentMetadata
	Id       int64 `json:"id"`
	Content  string
	Children *[]Comment
}

// TODO: Add report status
type Report struct {
	Id      int64
	Post    *Post
	Reason  string
	Creator *DisplayableUser
}
