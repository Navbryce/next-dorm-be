package db

import (
	"context"
	"database/sql"
	"github.com/navbryce/next-dorm-be/model"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type IntFilter struct {
	Val int64
}

type Database interface {
	CommunityDatabase
	PostDatabase
	SubscriptionDatabase
	UserDatabase
	GetSQLDB() *sql.DB
	Close() error
}

type GetCommunitiesQueryOpts struct {
	ForUserId string // will return subscription if it exists for user
}

type CommunityDatabase interface {
	CreateCommunity(ctx context.Context, name string) (communityId int64, err error)
	GetCommunitiesByIds(ctx context.Context, id []int64, opts *GetCommunitiesQueryOpts) ([]*model.CommunityWithSubStatus, error)
}

type CreateContentMetadata struct {
	CreatorId      string
	Visibility     model.Visibility
	CreatorAlias   string // only required if visibility is None
	ImageBlobNames []string
}

type EditContentMetadata struct {
	CreatorAlias           string
	ImageBlobNamesToAdd    []string
	ImageBlobNamesToRemove []string
	Visibility             model.Visibility
}

type CreatePost struct {
	*CreateContentMetadata
	Title       string
	Content     string
	Communities []int64
}

type EditPost struct {
	*EditContentMetadata
	Title   string
	Content string
}

type CreateComment struct {
	*CreateContentMetadata
	PostMetadataId   int64
	ParentMetadataId int64
	Content          string
}

type EditComment struct {
	*EditContentMetadata
	Content string
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
	CommunityIds   []int64
	IncludeDeleted bool
	*ByUser
	Visibility *model.Visibility
	PageByDate *ByDatePaging
	PageByVote *ByVotePaging
	*PostsListQueryOpts
}

type ByDatePaging struct {
	From   *time.Time
	LastId string // TODO: Change to int64
}

type ByVotePaging struct {
	MaxUpvotes *IntFilter
	Since      *time.Time
	LastId     string
}

type PostsListQueryOpts struct {
	Limit         int16
	VoteHistoryOf string
}

type CommentTreeQueryOpts struct {
	VoteHistoryOf string
}

type PostDatabase interface {
	CreatePost(ctx context.Context, req *CreatePost) (postId int64, err error)
	EditPost(ctx context.Context, id int64, req *EditPost) error
	CreateComment(ctx context.Context, req *CreateComment) (commentId int64, err error)
	EditComment(ctx context.Context, id int64, req *EditComment) error
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
	CreateUser(context.Context, *model.LocalUser) error
	GetUser(context.Context, string) (*model.LocalUser, error)
}
