package app

import (
	"context"
	"github.com/navbryce/next-dorm-be/model"
)

type RawCursor = map[string]interface{}

type PostCursorOpts struct {
	Limit int16
}

// TODO: Go generics?
type PostCursor interface {
	Posts(ctx context.Context, opts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error)
}

type PostCursorType string

const (
	PostCursorTypeMostRecent PostCursorType = "MOST_RECENT"
)
