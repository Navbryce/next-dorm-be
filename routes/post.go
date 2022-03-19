package routes

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type postRoutes struct {
	db db.Database
}

func AddPostRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := postRoutes{db}
	posts := group.Group("/posts", middleware.Auth(db, authClient, &middleware.AuthConfig{}))
	posts.GET("", util.HandlerWrapper(routes.getPosts, &util.HandlerOpts{}))
	posts.PUT("", util.HandlerWrapper(routes.createPost, &util.HandlerOpts{}))
	posts.GET("/:id", util.HandlerWrapper(routes.getPostById, &util.HandlerOpts{}))
	posts.DELETE("/:id", util.HandlerWrapper(routes.deletePost, &util.HandlerOpts{}))
	posts.PUT("/:id/votes", util.HandlerWrapper(routes.voteForPost, &util.HandlerOpts{}))
	posts.PUT("/:id/comments", util.HandlerWrapper(routes.createComment, &util.HandlerOpts{}))
	posts.GET("/:id/comments", util.HandlerWrapper(routes.getComments, &util.HandlerOpts{}))
	posts.PUT("/:id/comments/:comment-id/votes", util.HandlerWrapper(routes.voteForComment, &util.HandlerOpts{}))
	posts.PUT("/:id/reports", util.HandlerWrapper(routes.report, &util.HandlerOpts{}))
}

type createPostReq struct {
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Communities []int64          `json:"communities"`
	Visibility  model.Visibility `json:"visibility"`
}

func (pr *postRoutes) createPost(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createPostReq
	// TODO: Add validation
	// TODO: Auth by community
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	if len(req.Title) == 0 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "post must have title",
		}
	}

	if len(req.Content) == 0 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "post must have content",
		}
	}

	if len(req.Communities) == 0 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "post must belong to at least one community",
		}
	}

	communities, err := pr.db.GetCommunities(c, req.Communities)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	if len(communities) != len(req.Communities) {
		return nil, util.BuildDoesNotExistHTTPErr("community")
	}

	creatorAlias := ""
	if req.Visibility == model.VisibilityHidden {
		creatorAlias = util.GenerateAlias()
	}

	id, err := pr.db.CreatePost(c, &db.CreatePost{
		Title:       req.Title,
		Content:     req.Content,
		Communities: req.Communities,
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:    middleware.GetToken(c).UID,
			Visibility:   req.Visibility,
			CreatorAlias: creatorAlias,
		},
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return gin.H{
		"id": id,
	}, nil
}

func (pr *postRoutes) deletePost(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	if post.CanDelete(middleware.GetToken(c).UID) {
		return nil, util.BuildOperationForbidden("User is not the owner of the post")
	}
	if err := pr.db.MarkPostAsDeleted(c, post.Id); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return nil, nil
}

type createCommentReq struct {
	ParentCommentId int64            `json:"parentCommentId"`
	Content         string           `json:"content"`
	Visibility      model.Visibility `json:"visibility"`
}

func (pr *postRoutes) createComment(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createCommentReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	if len(req.Content) == 0 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "comment must have content",
		}
	}

	var parentMetadataId int64
	var rootMetadataId int64
	var httpErr *util.HTTPError
	if req.ParentCommentId == 0 {
		var post *model.Post
		post, httpErr = pr.mustGetPostByIdStr(c, c.Param("id"))
		if httpErr != nil {
			return nil, httpErr
		}
		rootMetadataId = post.ContentMetadata.Id
		parentMetadataId = rootMetadataId
	} else {
		comment, err := pr.db.GetCommentById(c, req.ParentCommentId)
		if err != nil {
			return nil, util.BuildDbHTTPErr(err)
		} else if comment == nil {
			return nil, util.BuildDoesNotExistHTTPErr("comment")
		}
		rootMetadataId = comment.RootMetadataId
		parentMetadataId = comment.ContentMetadata.Id
	}

	creatorAlias := ""
	if req.Visibility == model.VisibilityHidden {
		creatorAlias = util.GenerateAlias()
	}

	id, err := pr.db.CreateComment(c, &db.CreateComment{
		Content:          req.Content,
		RootMetadataId:   rootMetadataId,
		ParentMetadataId: parentMetadataId,
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:    middleware.GetToken(c).UID,
			Visibility:   req.Visibility,
			CreatorAlias: creatorAlias,
		},
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return &gin.H{
		"id": id,
	}, nil
}

func (pr *postRoutes) deleteComment(c *gin.Context) (interface{}, *util.HTTPError) {
	comment, httpErr := pr.mustGetCommentByIdStr(c, c.Param("comment-id"))
	if httpErr != nil {
		return nil, httpErr
	}
}

func (pr *postRoutes) getPostById(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	return post.MakeDisplayableFor(middleware.GetToken(c).UID), nil
}

func (pr *postRoutes) getPosts(c *gin.Context) (interface{}, *util.HTTPError) {
	var from *time.Time
	if c.Query("from") != "" {
		fromTime, err := time.Parse(time.RFC3339, c.Query("from"))
		if err != nil {
			return nil, &util.HTTPError{
				Status:  http.StatusBadRequest,
				Message: "Invalid date format",
			}
		}
		from = &fromTime
	}

	var communityIds []int64 = nil
	if c.Query("community") != "" {
		communityIdStrings := strings.Split(c.Query("community"), ",")
		communityIds = make([]int64, len(communityIdStrings))
		for i, communityIdString := range communityIdStrings {
			communityId, err := strconv.ParseInt(communityIdString, 10, 64)
			if err != nil {
				return nil, util.MalformedIdHTTPErr
			}
			communityIds[i] = communityId
		}
	}

	limit := int64(5)
	if c.Query("limit") != "" {
		var err error
		limit, err = strconv.ParseInt(c.Query("limit"), 10, 16)
		if err != nil {
			return nil, &util.HTTPError{
				Status:  http.StatusBadRequest,
				Message: "malformed limit",
			}
		}
		if limit > 500 {
			limit = 500
		}
	}
	cursor := c.Query("cursor")

	posts, err := pr.db.GetPosts(c, &db.PostsListQuery{
		From:         from,
		Cursor:       cursor,
		CommunityIds: communityIds,
		PostsListQueryOpts: &db.PostsListQueryOpts{
			Limit:         int16(limit),
			VoteHistoryOf: middleware.GetToken(c).UID,
		},
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	displayablePosts := make([]*model.Post, len(posts))
	for i, post := range posts {
		displayablePosts[i] = post.MakeDisplayableFor(middleware.GetToken(c).UID)
	}

	return displayablePosts, nil
}

func (pr *postRoutes) getComments(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}

	comments, err := pr.db.GetCommentForest(c, post.ContentMetadata.Id, &db.CommentTreeQueryOpts{VoteHistoryOf: middleware.GetToken(c).UID})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}

	for i, comment := range comments {
		comments[i] = comment.MakeDisplayableFor(middleware.GetToken(c).UID)
	}

	return comments, nil
}

type voteReq struct {
	Value int8
}

func (pr *postRoutes) voteForPost(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}

	var req voteReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	if err := pr.db.Vote(c, middleware.GetToken(c).UID, post.ContentMetadata.Id, normalizeVote(req.Value)); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return nil, nil
}

func (pr *postRoutes) voteForComment(c *gin.Context) (interface{}, *util.HTTPError) {
	comment, httpErr := pr.mustGetCommentByIdStr(c, c.Param("comment-id"))
	if httpErr != nil {
		return nil, httpErr
	}

	if strconv.FormatInt(comment.RootMetadataId, 10) != c.Param("id") {
		return nil, &util.HTTPError{
			Status:  http.StatusNotFound,
			Message: "comment does not exist under post",
		}
	}

	var req voteReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	if err := pr.db.Vote(c, middleware.GetToken(c).UID, comment.ContentMetadata.Id, normalizeVote(req.Value)); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return nil, nil
}

func normalizeVote(value int8) int8 {
	if value < -1 {
		return -1
	} else if value > 1 {
		return 1
	}
	return value

}

type reportReq struct {
	Reason string
}

// TODO: Add GetReportByPost and GetReportByCommunity
func (pr *postRoutes) report(c *gin.Context) (interface{}, *util.HTTPError) {
	var req reportReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	/* TODO: Possible race condition if post was deleted
	right after report created
	*/
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}

	reportId, err := pr.db.CreateReport(c, middleware.GetToken(c).UID, &db.CreateReport{
		PostId: post.Id,
		Reason: req.Reason,
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return gin.H{
		"id": reportId,
	}, nil
}

// mustGetPostByIdStr attempts to get post by id str
func (pr *postRoutes) mustGetPostByIdStr(ctx context.Context, idStr string) (*model.Post, *util.HTTPError) {
	if post, err := mustGetByIdStr(ctx, func(ctx context.Context, id int64) (entity interface{}, isNil bool, dbErr error) {
		post, err := pr.db.GetPostById(ctx, id, &db.PostQueryOpts{})
		return post, post == nil, err
	}, "post", idStr); err != nil {
		return nil, err
	} else {
		return post.(*model.Post), nil
	}

}

// mustGetCommentByIdStr attempts to get post by id str
func (pr *postRoutes) mustGetCommentByIdStr(ctx context.Context, idStr string) (*model.Comment, *util.HTTPError) {
	if post, err := mustGetByIdStr(ctx, func(ctx context.Context, id int64) (entity interface{}, isNil bool, dbErr error) {
		post, err := pr.db.GetCommentById(ctx, id)
		return post, post == nil, err
	}, "comment", idStr); err != nil {
		return nil, err
	} else {
		return post.(*model.Comment), nil
	}

}

type FetchById = func(ctx context.Context, id int64) (entity interface{}, isNil bool, dbErr error)

func mustGetByIdStr(ctx context.Context, fetch FetchById, entityType string, idStr string) (interface{}, *util.HTTPError) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, util.MalformedIdHTTPErr
	}
	entity, isNil, err := fetch(ctx, id)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	} else if isNil {
		return nil, util.BuildDoesNotExistHTTPErr(entityType)
	}
	return entity, nil
}
