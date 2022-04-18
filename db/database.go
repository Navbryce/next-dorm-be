package db

import (
	"context"
	"database/sql"
	"github.com/navbryce/next-dorm-be/model"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	PostDatabase
	SubscriptionDatabase
	UserDatabase
	GetSQLDB() *sql.DB
	Close() error
}

type GetCommunitiesQueryOpts struct {
	ForUserId string // will return subscription if it exists for user
}

type CreateContentMetadata struct {
	CreatorId      string
	Visibility     model.Visibility
	CreatorAlias   string // only required if visibility is None
	ImageBlobNames []string
}

type CreatePost struct {
	*CreateContentMetadata
	Title       string
	Content     string
	Communities []int64
}

type CreateComment struct {
	*CreateContentMetadata
	RootMetadataId   int64
	ParentMetadataId int64
	Content          string
}

type CreateReport struct {
	PostId int64
	Reason string
}

type PostQueryOpts struct {
	VoteHistoryOf string
}

type ByUser struct {
	Id string
}

type PostsListQuery struct {
	From         *time.Time
	LastId       string // TODO: Change to int64
	CommunityIds []int64
	*ByUser
	model.Visibility
	*PostsListQueryOpts
}

type PostsListQueryOpts struct {
	Limit         int16
	VoteHistoryOf string
}

type CommentTreeQueryOpts struct {
	VoteHistoryOf string
}

type PostDatabase interface {
	CreateCommunity(ctx context.Context, name string) (communityId int64, err error)
	GetCommunities(ctx context.Context, ids []int64, opts *GetCommunitiesQueryOpts) ([]*model.CommunityWithSubStatus, error)
	CreatePost(ctx context.Context, req *CreatePost) (postId int64, err error)
	CreateComment(ctx context.Context, req *CreateComment) (commentId int64, err error)
	MarkPostAsDeleted(context.Context, int64) error
	MarkCommentAsDeleted(context.Context, int64) error
	GetPostById(context.Context, int64, *PostQueryOpts) (*model.Post, error)
	GetPosts(context.Context, *PostsListQuery) ([]*model.Post, error)
	GetCommentById(ctx context.Context, id int64) (*model.Comment, error)
	GetCommentForest(ctx context.Context, rootMetadataId int64, opts *CommentTreeQueryOpts) ([]*model.CommentTree, error)
	Vote(ctx context.Context, userId string, contentMetadataId int64, value int8) error
	CreateReport(ctx context.Context, userId string, req *CreateReport) (reportId int64, err error)
}

type SubscriptionDatabase interface {
	CreateSubForUser(context.Context, *model.Subscription) error
	GetSubsForUser(ctx context.Context, userId string) ([]*model.Subscription, error)
	DeleteSubForUser(context.Context, *model.Subscription) error
}

type UserDatabase interface {
	CreateUser(context.Context, *model.User) error
	GetUser(context.Context, string) (*model.User, error)
}
