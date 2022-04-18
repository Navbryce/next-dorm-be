package model

import (
	"time"
)

type Visibility string

const (
	VisibilityNormal Visibility = "NORMAL"
	VisibilityHidden            = "HIDDEN"
)

type Status string

const (
	StatusPosted  Status = "POSTED"
	StatusDeleted        = "DELETED"
)

type Vote struct {
	Value int8 `json:"value"`
}

type ContentMetadata struct {
	Id             int64            `json:"-"`
	Creator        *DisplayableUser `json:"creator"`
	UserVote       *Vote            `json:"userVote"`
	Status         `json:"status"`
	Visibility     Visibility `json:"visibility"`
	NumVotes       int        `json:"numVotes"`
	VoteTotal      int        `json:"voteTotal"`
	ImageBlobNames []string   `json:"imageBlobNames"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (cm *ContentMetadata) MakeDisplayableFor(user *User) *ContentMetadata {
	// TODO: Refactor into method checking if user is the person or an admin
	if user != nil && (user.IsAdmin || user.Id == cm.Creator.Id) {
		return cm
	}

	switch cm.Visibility {
	case VisibilityHidden:
		cm.Creator = &DisplayableUser{AnonymousUser: cm.Creator.AnonymousUser}
	case VisibilityNormal:
		cm.Creator = &DisplayableUser{User: cm.Creator.User.MakeDisplayableFor(user)}
	}

	return cm
}

func (cm *ContentMetadata) CanDelete(user *User) bool {
	return user.Id == cm.Creator.Id
}

type Post struct {
	*ContentMetadata
	Id          int64        `json:"id"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Communities []*Community `json:"communities"`
}

// MakeDisplayableFor mutates the object
func (p *Post) MakeDisplayableFor(user *User) *Post {
	p.ContentMetadata = p.ContentMetadata.MakeDisplayableFor(user)
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
func (ct *CommentTree) MakeDisplayableFor(user *User) *CommentTree {
	ct.ContentMetadata = ct.ContentMetadata.MakeDisplayableFor(user)
	for i, child := range ct.Children {
		ct.Children[i] = child.MakeDisplayableFor(user)
		child.ContentMetadata = child.ContentMetadata.MakeDisplayableFor(user)
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
