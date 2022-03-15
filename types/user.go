package types

// User holds the local user data relevant to the application (outside of frirebase)
type User struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
}

type DisplayableUser struct {
	*User
	Alias string `json:"alias,omitempty"`
}
