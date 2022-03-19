package model

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

type Vote struct {
	Value int8 `json:"value"`
}

type ContentMetadata struct {
	Id         int64            `json:"-"`
	Creator    *DisplayableUser `json:"creator"`
	UserVote   *Vote            `json:"userVote"`
	Visibility Visibility       `json:"visibility"`
	NumVotes   int              `json:"numVotes"`
	VoteTotal  int              `json:"voteTotal"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

func (cm *ContentMetadata) MakeDisplayableFor(userId string) *ContentMetadata {
	if userId == cm.Creator.Id {
		return cm
	}

	switch cm.Visibility {
	case VisibilityHidden:
		cm.Creator = &DisplayableUser{Alias: cm.Creator.Alias}
	}

	return cm
}

func (cm *ContentMetadata) CanDelete(userId string) bool {
	return userId == cm.Creator.Id
}

type Post struct {
	*ContentMetadata
	Id          int64        `json:"id"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Communities []*Community `json:"communities"`
}

// MakeDisplayableFor mutates the object
func (p *Post) MakeDisplayableFor(userId string) *Post {
	p.ContentMetadata = p.ContentMetadata.MakeDisplayableFor(userId)
	return p
}

type Comment struct {
	*ContentMetadata
	Id               int64  `json:"id"`
	ParentMetadataId int64  `json:"-"`
	RootMetadataId   int64  `json:"-"`
	Content          string `json:"content"`
}

type CommentTree struct {
	*Comment
	Children []*CommentTree `json:"children"`
}

// MakeDisplayableFor mutates the object
func (ct *CommentTree) MakeDisplayableFor(userId string) *CommentTree {
	for i, tree := range ct.Children {
		ct.Children[i] = tree.MakeDisplayableFor(userId)
		tree.ContentMetadata = tree.ContentMetadata.MakeDisplayableFor(userId)
	}
	return ct
}

// TODO: Add report status
type Report struct {
	Id      int64
	Post    *Post
	Reason  string
	Creator *DisplayableUser
}
