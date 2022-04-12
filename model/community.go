package model

type Community struct {
	Id        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ParentId  int64  `db:"parent_id" json:"parentId"`
	HasParent bool   `db:"has_parent" json:"hasParent"` // TODO: Add has parent
}

type CommunityWithSubStatus struct {
	*Community
	// TODO: Just change to subscription status?
	IsSubscribed bool `db:"is_subscribed" json:"isSubscribed"`
}
