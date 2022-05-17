package model

import (
	"github.com/navbryce/next-dorm-be/db/dao"
	"time"
)

type Community struct {
	Id        int64         `db:"id" json:"id"`
	Name      string        `db:"name" json:"name"`
	ParentId  dao.NullInt64 `db:"parent_id" json:"parentId"`
	CreatedAt *time.Time    `db:"created_at"`
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
