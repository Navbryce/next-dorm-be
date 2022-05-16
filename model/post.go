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
	Id             int64          `json:"-"`
	Creator        *ContentAuthor `json:"creator"`
	UserVote       *Vote          `json:"userVote"`
	Status         `json:"status"`
	Visibility     Visibility `json:"visibility"`
	NumVotes       uint64     `json:"numVotes"`
	VoteTotal      int64      `json:"voteTotal"`
	ImageBlobNames []string   `json:"imageBlobNames"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (cm *ContentMetadata) MakeDisplayableFor(user *LocalUser) *ContentMetadata {
	switch cm.Visibility {
	case VisibilityHidden:
		// TODO: Refactor into method checking if user is the person or an admin
		if user != nil && (user.IsAdmin || user.Id == cm.Creator.Id) {
			return cm
		}
		cm.Creator = &ContentAuthor{AnonymousUser: cm.Creator.AnonymousUser}
	case VisibilityNormal:
		cm.Creator = &ContentAuthor{LocalUser: cm.Creator.LocalUser.MakeDisplayableFor(user)}
	}

	return cm
}

// TODO: move this logic into controller?
func (cm *ContentMetadata) CanEdit(user *LocalUser) bool {
	if cm.Status == StatusDeleted {
		return false
	}
	return user.Id == cm.Creator.Id || user.IsAdmin
}

// TODO: Move this logic into controller?
func (cm *ContentMetadata) CanDelete(user *LocalUser) bool {
	return user.Id == cm.Creator.Id || user.IsAdmin
}

type Post struct {
	*ContentMetadata
	Id           int64        `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	Communities  []*Community `json:"communities"`
	CommentCount int64        `json:"commentCount"`
}

// MakeDisplayableFor mutates the object
func (p *Post) MakeDisplayableFor(user *LocalUser) *Post {
	p.ContentMetadata = p.ContentMetadata.MakeDisplayableFor(user)
	return p
}

func MakePostsDisplayableFor(posts []*Post, user *LocalUser) []*Post {
	displayablePosts := make([]*Post, len(posts))
	for i, post := range posts {
		displayablePosts[i] = post.MakeDisplayableFor(user)
	}
	return displayablePosts
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
func (ct *CommentTree) MakeDisplayableFor(user *LocalUser) *CommentTree {
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
	Creator *ContentAuthor
}
