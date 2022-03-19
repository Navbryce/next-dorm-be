package model

// User holds the local user data relevant to the application (outside of firebase db)
type User struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
	IsAdmin     bool   `db:"isAdmin" json:"isAdmin"`
}

type DisplayableUser struct {
	*User
	Alias string `json:"alias,omitempty"`
}
