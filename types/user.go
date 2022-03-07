package types

// User holds the local user data relevant to the application
type User struct {
	Id          string `db:"firebase_id"`
	DisplayName string `db:"display_name"`
}

type DisplayableUser = User
