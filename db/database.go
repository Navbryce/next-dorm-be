package db

import (
	"context"
	"database/sql"
	"github.com/heroku/go-getting-started/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type CreatePost struct {
	Content     string
	Communities []int64
	Visibility  types.Visibility
}

type Database interface {
	PostDatabase
	GetSQLDB() *sql.DB
	Close() error
}

type PostDatabase interface {
	CreateCommunity(ctx context.Context, name string) (communityId int64, err error)
	CreatePost(context.Context, *CreatePost) (postId int64, err error)
	GetPostById(ctx context.Context, id int64) (*types.Post, error)
	GetPosts(ctx context.Context, from time.Time, cursor string, communityIds []int64, limit int16) ([]*types.Post, error)
	GetCommunities(context.Context) ([]*types.Community, error)
}
