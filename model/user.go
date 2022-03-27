package model

// User holds the local user data relevant to the application (outside of firebase db)
type User struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
	IsAdmin     bool   `db:"isAdmin" json:"isAdmin"`
	Avatar      string `json:"avatar"`
}

// TODO: Separate DAO and data classes
// TODO: Create ContentAuthor struct with a MakeDisplayable

type AnonymousUser struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type DisplayableUser struct {
	*AnonymousUser `json:"anonymousUser,omitempty"`
	*User          `json:"user,omitempty"`
}
