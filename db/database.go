package db

import (
	"context"
	"database/sql"
	"github.com/heroku/go-getting-started/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	PostDatabase
	UserDatabase
	GetSQLDB() *sql.DB
	Close() error
}

type CreatePost struct {
	Content     string
	Communities []int64
	Visibility  types.Visibility
}

type PostDatabase interface {
	CreateCommunity(ctx context.Context, name string) (communityId int64, err error)
	GetCommunities(ctx context.Context, ids []int64) ([]*types.Community, error)
	CreatePost(context.Context, string, *CreatePost) (postId int64, err error)
	GetPostById(ctx context.Context, id int64) (*types.Post, error)
	GetPosts(ctx context.Context, from *time.Time, cursor string, communityIds []int64, limit int16) ([]*types.Post, error)
	Vote(ctx context.Context, userId string, postId int64, value int8) error
}

type UserDatabase interface {
	CreateUser(context.Context, *types.User) error
	GetUser(context.Context, string) (*types.User, error)
}
