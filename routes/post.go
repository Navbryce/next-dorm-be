package routes

import (
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/app"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/services"
	"github.com/navbryce/next-dorm-be/util"
	"log"
	"net/http"
	"strconv"
)

type postRoutes struct {
	db                db.Database
	userUploadsBucket *services.StorageBucket
}

func AddPostRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client, userUploadsBucket *services.StorageBucket) {
	routes := postRoutes{db, userUploadsBucket}
	posts := group.Group("/posts", middleware.GenAuth(db, authClient, &middleware.AuthConfig{}))
	posts.POST("",
		util.HandlerWrapper(routes.getPosts, &util.HandlerOpts{}))
	posts.PUT("", middleware.RequireAccount(), util.HandlerWrapper(routes.createPost, &util.HandlerOpts{}))
	posts.GET("/:id", util.HandlerWrapper(routes.getPostById, &util.HandlerOpts{}))
	posts.PUT("/:id", middleware.RequireAccount(), util.HandlerWrapper(routes.editPost, &util.HandlerOpts{}))
	posts.DELETE("/:id", middleware.RequireAccount(), util.HandlerWrapper(routes.deletePost, &util.HandlerOpts{}))
	posts.PUT("/:id/votes", middleware.RequireAccount(), util.HandlerWrapper(routes.voteForPost, &util.HandlerOpts{}))
	posts.PUT("/:id/comments", middleware.RequireAccount(), util.HandlerWrapper(routes.createComment, &util.HandlerOpts{}))
	posts.GET("/:id/comments", util.HandlerWrapper(routes.getComments, &util.HandlerOpts{}))
	posts.PUT("/:id/comments/:comment-id", middleware.RequireAccount(), util.HandlerWrapper(routes.editComment, &util.HandlerOpts{}))
	posts.DELETE("/:id/comments/:comment-id", middleware.RequireAccount(), util.HandlerWrapper(routes.deleteComment, &util.HandlerOpts{}))
	posts.PUT("/:id/comments/:comment-id/votes", middleware.RequireAccount(), util.HandlerWrapper(routes.voteForComment, &util.HandlerOpts{}))
	posts.PUT("/:id/reports", middleware.RequireAccount(), util.HandlerWrapper(routes.report, &util.HandlerOpts{}))
}

type createPostReq struct {
	Title          string           `json:"title"`
	Content        string           `json:"content"`
	Communities    []int64          `json:"communities"`
	Visibility     model.Visibility `json:"visibility"`
	ImageBlobNames []string         `json:"imageBlobNames"`
}

func (pr *postRoutes) createPost(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createPostReq
	// TODO: Add validation
	// TODO: GenAuth by community
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

	// TODO: Enable multiple communities in the future?
	if len(req.Communities) != 1 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "post must belong to at exactly one community",
		}
	}

	communities, err := pr.db.GetCommunitiesByIds(c, req.Communities, &db.GetCommunitiesQueryOpts{})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	if len(communities) != len(req.Communities) {
		return nil, util.BuildDoesNotExistHTTPErr("community")
	}

	if err := pr.imagesMustExist(c, req.ImageBlobNames); err != nil {
		return nil, err
	}

	var creatorAlias model.AnonymousUser
	if req.Visibility == model.VisibilityHidden {
		creatorAlias = *util.GenerateAnonymousUser()
	}

	id, err := pr.db.CreatePost(c, &db.CreatePost{
		Title:       req.Title,
		Content:     req.Content,
		Communities: req.Communities,
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:      middleware.MustGetToken(c).UID,
			Visibility:     req.Visibility,
			CreatorAlias:   creatorAlias.DisplayName,
			ImageBlobNames: req.ImageBlobNames,
		},
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return gin.H{
		"id": id,
	}, nil
}

type editPostImages struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

type editPostReq struct {
	// TODO: Change all fields to dif. Not just images
	Title          string           `json:"title"`
	Content        string           `json:"content"`
	ImageBlobNames editPostImages   `json:"imageBlobNames"`
	Visibility     model.Visibility `json:"visibility"`
}

func (pr *postRoutes) editPost(c *gin.Context) (interface{}, *util.HTTPError) {
	// TODO: ADD OWNERSHIP CHECK FOR EDITING POST
	// TODO: Optimize req to only send fields that are changed. send a diff
	var req editPostReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	post, err := pr.mustGetPostByIdStr(c, c.Param("id"))
	if err != nil {
		return nil, err
	}

	if !post.CanEdit(middleware.GetUser(c)) {
		// TODO: Create permission checking system where model just defines a permission object
		return nil, util.BuildOperationForbidden("must be owner or admin. or the content is deleted")
	}

	if err := pr.imagesMustExist(c, req.ImageBlobNames.Added); err != nil {
		return nil, err
	}

	var newAlias model.AnonymousUser
	if len(post.Creator.AnonymousUser.DisplayName) == 0 && req.Visibility == model.VisibilityHidden {
		newAlias = *util.GenerateAnonymousUser()
	}

	if err := pr.db.EditPost(c, post.Id, &db.EditPost{
		Title:   req.Title,
		Content: req.Content,
		EditContentMetadata: &db.EditContentMetadata{
			ImageBlobNamesToAdd:    req.ImageBlobNames.Added,
			ImageBlobNamesToRemove: req.ImageBlobNames.Removed,
			Visibility:             req.Visibility,
			CreatorAlias:           newAlias.DisplayName,
		},
	},
	); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return nil, nil
}

// TODO: Move logic to controllers
func (pr *postRoutes) deletePost(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	if !post.CanDelete(middleware.MustGetUser(c)) {
		// TODO: Switch to permission system
		return nil, util.BuildOperationForbidden("user is not the owner of the post or an admin")
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

	var alias *model.AnonymousUser
	aliasDisplayName := ""
	if req.Visibility == model.VisibilityHidden {
		alias = util.GenerateAnonymousUser()
		aliasDisplayName = alias.DisplayName
	}

	id, err := pr.db.CreateComment(c, &db.CreateComment{
		Content:          req.Content,
		RootMetadataId:   rootMetadataId,   // TODO: Switch to post id?
		ParentMetadataId: parentMetadataId, // TODO: Switch to parent comment id?
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:    middleware.MustGetToken(c).UID,
			Visibility:   req.Visibility,
			CreatorAlias: aliasDisplayName,
		},
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return &gin.H{
		"id":    id,
		"alias": alias,
	}, nil
}

// TODO: Factor out editMetadata into its own struct?
type editCommentReq struct {
	Content    string           `json:"content"`
	Visibility model.Visibility `json:"visibility"`
}

func (pr *postRoutes) editComment(c *gin.Context) (interface{}, *util.HTTPError) {
	// TODO: We're not using post id. Keep it?
	var req editCommentReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	comment, httpErr := pr.mustGetCommentByIdStr(c, c.Param("comment-id"))
	if httpErr != nil {
		return nil, httpErr
	}
	// TODO: Check if comment exists under post?
	if !comment.CanEdit(middleware.MustGetUser(c)) {
		return nil, util.BuildOperationForbidden("user is not owner of the comment or admin. or the content is deleted.")
	}

	newAliasDisplayName := ""
	var alias *model.AnonymousUser
	if req.Visibility == model.VisibilityHidden {
		alias = comment.Creator.AnonymousUser
		if len(comment.Creator.AnonymousUser.DisplayName) == 0 && req.Visibility == model.VisibilityHidden {
			alias = util.GenerateAnonymousUser()
			newAliasDisplayName = alias.DisplayName
		}
	}

	if err := pr.db.EditComment(c, comment.Id, &db.EditComment{
		EditContentMetadata: &db.EditContentMetadata{
			Visibility:   req.Visibility,
			CreatorAlias: newAliasDisplayName,
		},
		Content: req.Content,
	}); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}

	return gin.H{"alias": alias}, nil
}

func (pr *postRoutes) deleteComment(c *gin.Context) (interface{}, *util.HTTPError) {
	comment, httpErr := pr.mustGetCommentByIdStr(c, c.Param("comment-id"))
	if httpErr != nil {
		return nil, httpErr
	}
	// TODO: Check if comment exists under post?
	if !comment.CanDelete(middleware.MustGetUser(c)) {
		return nil, util.BuildOperationForbidden("user is not owner of the post")
	}
	if err := pr.db.MarkCommentAsDeleted(c, comment.Id); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return nil, nil
}

func (pr *postRoutes) getPostById(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	return post.MakeDisplayableFor(middleware.GetUser(c)), nil
}

// TODO: Turn cursor into struct with fields for each type and add methods for each type. No enum
type getPostsReq struct {
	app.TaggedUnionCursor
}

func (pr *postRoutes) getPosts(c *gin.Context) (interface{}, *util.HTTPError) {
	var req getPostsReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	cursor := req.PostCursor
	switch v := cursor.(type) {
	case *app.MostRecentCursor:
		if !canViewHiddenPostsByUser(middleware.GetUser(c), v.ByUser) {
			visibility := model.VisibilityNormal
			v.Visibility = &visibility
		}
	case *app.MostPopularCursor:
		if !canViewHiddenPostsByUser(middleware.GetUser(c), v.ByUser) {
			visibility := model.VisibilityNormal
			v.Visibility = &visibility
		}
	}
	posts, nextCursor, err := cursor.Posts(c, pr.db, middleware.GetUser(c), &app.PostCursorOpts{Limit: 20})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}

	return gin.H{
		"posts":      model.MakePostsDisplayableFor(posts, middleware.GetUser(c)),
		"nextCursor": nextCursor,
	}, nil
}

func canViewHiddenPostsByUser(userMaybe *model.User, byUserMaybe *app.SerializableByUser) bool {
	return byUserMaybe == nil ||
		(userMaybe != nil && (userMaybe.Id == byUserMaybe.Id || userMaybe.IsAdmin))
}

func (pr *postRoutes) getComments(c *gin.Context) (interface{}, *util.HTTPError) {
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	voteHistoryOf := ""
	if middleware.GetToken(c) != nil {
		voteHistoryOf = middleware.GetToken(c).UID
	}
	comments, err := pr.db.GetCommentForest(c, post.ContentMetadata.Id, &db.CommentTreeQueryOpts{VoteHistoryOf: voteHistoryOf})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}

	for i, comment := range comments {
		comments[i] = comment.MakeDisplayableFor(middleware.GetUser(c))
	}

	return comments, nil
}

type voteReq struct {
	Value int8 `json:"value"`
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
	if err := pr.db.Vote(c, middleware.MustGetToken(c).UID, post.ContentMetadata.Id, normalizeVote(req.Value)); err != nil {
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

	if err := pr.db.Vote(c, middleware.MustGetToken(c).UID, comment.ContentMetadata.Id, normalizeVote(req.Value)); err != nil {
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

	reportId, err := pr.db.CreateReport(c, middleware.MustGetToken(c).UID, &db.CreateReport{
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
func (pr *postRoutes) mustGetPostByIdStr(ctx *gin.Context, idStr string) (*model.Post, *util.HTTPError) {
	if post, err := mustGetByIdStr(ctx, func(ctx *gin.Context, id int64) (entity interface{}, isNil bool, dbErr error) {
		post, err := pr.db.GetPostById(ctx, id, &db.PostQueryOpts{
			VoteHistoryOf: middleware.GetUserIdMaybe(ctx),
		})
		return post, post == nil, err
	}, "post", idStr); err != nil {
		return nil, err
	} else {
		return post.(*model.Post), nil
	}

}

// mustGetCommentByIdStr attempts to get post by id str
func (pr *postRoutes) mustGetCommentByIdStr(ctx *gin.Context, idStr string) (*model.Comment, *util.HTTPError) {
	if post, err := mustGetByIdStr(ctx, func(ctx *gin.Context, id int64) (entity interface{}, isNil bool, dbErr error) {
		post, err := pr.db.GetCommentById(ctx, id)
		return post, post == nil, err
	}, "comment", idStr); err != nil {
		return nil, err
	} else {
		return post.(*model.Comment), nil
	}

}

type FetchById = func(ctx *gin.Context, id int64) (entity interface{}, isNil bool, dbErr error)

func mustGetByIdStr(ctx *gin.Context, fetch FetchById, entityType string, idStr string) (interface{}, *util.HTTPError) {
	id, httpErr := util.ParseId(idStr)
	if httpErr != nil {
		return nil, httpErr
	}
	entity, isNil, err := fetch(ctx, id)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	} else if isNil {
		return nil, util.BuildDoesNotExistHTTPErr(entityType)
	}
	return entity, nil
}

func (pr *postRoutes) imagesMustExist(c *gin.Context, imageBlobNames []string) *util.HTTPError {
	for _, blobName := range imageBlobNames {
		if exists, err := pr.userUploadsBucket.Exists(c, blobName); err != nil {
			log.Println("a storage error occurred", err)
			return &util.HTTPError{
				Status:  http.StatusInternalServerError,
				Message: "a storage error occurred",
			}
		} else if !exists {
			return &util.HTTPError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("uploaded image does not exist %v", blobName),
			}
		}
	}
	return nil
}
