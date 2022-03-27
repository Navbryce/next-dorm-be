package model

type Subscription struct {
	UserId      string `db:"user_id" json:"userId"`
	CommunityId int64  `db:"community_id" json:"communityId"`
}
