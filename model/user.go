package model

import (
	"encoding/json"
	"firebase.google.com/go/v4/auth"
	"fmt"
)

func avatarBlobNameFromId(userId string) string {
	return fmt.Sprintf("uploads/%v/avatar", userId)
}

type FirebaseToken struct {
	*auth.Token
}

func (fbu *FirebaseToken) AvatarBlobNameForUser() string {
	return avatarBlobNameFromId(fbu.UID)
}

// LocalUser holds the local user data relevant to the application (outside of firebase db)
// TODO: Split DAO from app model
type LocalUser struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
	IsAdmin     bool   `db:"is_admin" json:"isAdmin"`
}

func (u *LocalUser) AvatarBlobNameForUser() string {
	return avatarBlobNameFromId(u.Id)
}

func (u *LocalUser) MakeDisplayableFor(user *LocalUser) *LocalUser {
	if user != nil && (user.IsAdmin || user.Id == u.Id) {
		return user
	}
	return &LocalUser{
		Id:          u.Id,
		DisplayName: u.DisplayName,
	}
}

func (u *LocalUser) MarshalJSON() ([]byte, error) {
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
	*LocalUser
	*AnonymousUser
}

// MarshalJSON prevents inheriting the *LocalUser MarshalJSON
func (c *ContentAuthor) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		User          *LocalUser     `json:"user,omitempty"`
		AnonymousUser *AnonymousUser `json:"anonymousUser,omitempty"`
	}{c.LocalUser, c.AnonymousUser})
}
