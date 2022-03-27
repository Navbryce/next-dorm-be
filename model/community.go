package model

type Community struct {
	Id        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ParentId  int64  `db:"parent_id" json:"parentId"`
	HasParent bool   `db:"has_parent" json:"hasParent"` // TODO: Add has parent
}
