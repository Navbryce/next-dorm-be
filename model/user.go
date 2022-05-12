package model

import (
	"encoding/json"
	"fmt"
)

// User holds the local user data relevant to the application (outside of firebase db)
type User struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
	IsAdmin     bool   `db:"is_admin" json:"isAdmin"`
}

func (u *User) AvatarBlobNameForUser() string {
	return fmt.Sprintf("avatars/%v", u.Id)
}

func (u *User) MakeDisplayableFor(user *User) *User {
	if user != nil && (user.IsAdmin || user.Id == u.Id) {
		return user
	}
	return &User{
		Id:          u.Id,
		DisplayName: u.DisplayName,
	}
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id             string `json:"id"`
		DisplayName    string `json:"displayName"`
		IsAdmin        bool   `json:"isAdmin"`
		AvatarBlobName string `json:"avatarBlobName"`
	}{
		Id:             u.Id,
		DisplayName:    u.DisplayName,
		IsAdmin:        u.IsAdmin,
		AvatarBlobName: u.AvatarBlobNameForUser(),
	})
}

// TODO: Separate DAO and data classes
//
//TODO: Create ContentAuthor struct with a MakeDisplayable. Make AnonymousUser just a field (alias) rather than a
// struct
type AnonymousUser struct {
	DisplayName string `json:"displayName"`
	AvatarUrl   string `json:"avatarUrl"`
}

// TODO: Rename: ContentAuthor?
type ContentAuthor struct {
	*User
	*AnonymousUser
}

// MarshalJSON prevents inheriting the *User MarshalJSON
func (c *ContentAuthor) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		User          *User          `json:"user,omitempty"`
		AnonymousUser *AnonymousUser `json:"anonymousUser,omitempty"`
	}{c.User, c.AnonymousUser})
}
