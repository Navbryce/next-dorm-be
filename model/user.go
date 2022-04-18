package model

// User holds the local user data relevant to the application (outside of firebase db)
type User struct {
	Id          string `db:"firebase_id" json:"id"`
	DisplayName string `db:"display_name" json:"displayName"`
	IsAdmin     bool   `db:"is_admin" json:"isAdmin"`
	Avatar      string `db:"avatar" json:"avatar"`
}

func (u *User) MakeDisplayableFor(user *User) *User {
	if user != nil && (user.IsAdmin || user.Id == u.Id) {
		return user
	}
	return &User{
		Id:          u.Id,
		DisplayName: u.DisplayName,
		Avatar:      u.Avatar,
	}
}

// TODO: Separate DAO and data classes
// TODO: Create ContentAuthor struct with a MakeDisplayable

type AnonymousUser struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

// TODO: Rename: ContentAuthor?
type DisplayableUser struct {
	*AnonymousUser `json:"anonymousUser,omitempty"`
	*User          `json:"user,omitempty"`
}
