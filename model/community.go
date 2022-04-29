package model

import "time"

type Community struct {
	Id   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	// TODO: If 0, has no parent. Change this to a nilable pointer
	ParentId  int64      `db:"parent_id" json:"parentId"`
	CreatedAt *time.Time `db:"created_at"`
}

type CommunityWithSubStatus struct {
	*Community
	// TODO: Just change to subscription status?
	IsSubscribed bool `db:"is_subscribed" json:"isSubscribed"`
}

type CommunityPosInTree struct {
	Children []*Community `json:"children"`
	Path     []*Community `json:"path"`
}
