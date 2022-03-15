package db

import (
	"context"
	"database/sql"
	"github.com/navbryce/next-dorm-be/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	PostDatabase
	UserDatabase
	GetSQLDB() *sql.DB
	Close() error
}

type CreateContentMetadata struct {
	CreatorId    string
	Visibility   types.Visibility
	CreatorAlias string // only required if visibility is None
}

type CreatePost struct {
	*CreateContentMetadata
	Title       string
	Content     string
	Communities []int64
}

type CreateComment struct {
	*CreateContentMetadata
	ParentMetadataId int64
	Content          string
}

type CreateReport struct {
	PostId int64
	Reason string
}

type PostDatabase interface {
	CreateCommunity(ctx context.Context, name string) (communityId int64, err error)
	GetCommunities(ctx context.Context, ids []int64) ([]*types.Community, error)
	CreatePost(ctx context.Context, req *CreatePost) (postId int64, err error)
	CreateComment(ctx context.Context, req *CreateComment) (commentId int64, err error)
	GetPostById(ctx context.Context, id int64) (*types.Post, error)
	GetPosts(ctx context.Context, from *time.Time, cursor string, communityIds []int64, limit int16) ([]*types.Post, error)
	//GetComments(ctx context.Context, from *time.Time, cursor string, limit int16) ([]*types.Post, error)
	Vote(ctx context.Context, userId string, contentMetadataId int64, value int8) error
	CreateReport(ctx context.Context, userId string, req *CreateReport) (reportId int64, err error)
}

type UserDatabase interface {
	CreateUser(context.Context, *types.User) error
	GetUser(context.Context, string) (*types.User, error)
}
