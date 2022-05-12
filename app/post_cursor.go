package app

import (
	"context"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
)

type PostCursorOpts struct {
	Limit int16
}

// TODO: Go generics?
type PostCursor interface {
	Posts(ctx context.Context, db appDb.Database, user *model.User, opts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error)
}

type PostCursorType string
