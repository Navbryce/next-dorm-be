package routes

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/types"
	"github.com/navbryce/next-dorm-be/util"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	postDoesNotExistHTTPErr = util.HTTPError{
		Message: "post not found",
		Status:  http.StatusNotFound,
	}
)

type postRoutes struct {
	db db.Database
}

func AddPostRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := postRoutes{db}
	posts := group.Group("/posts", Auth(db, authClient, &AuthConfig{}))
	posts.GET("", routes.getPosts)
	posts.PUT("", routes.createPost)
	posts.GET("/:id", routes.getPostById)
	posts.PUT("/:id/comment", routes.createComment)
	posts.PUT("/:id/vote", routes.vote)
	posts.PUT("/:id/report", routes.report)
}

type createPostReq struct {
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Communities []int64          `json:"communities"`
	Visibility  types.Visibility `json:"visibility"`
}

func (pr *postRoutes) createPost(c *gin.Context) {
	var req createPostReq
	// TODO: Add validation and standardize responses
	// TODO: Auth by community
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	if len(req.Title) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "post must have title",
		})
		return
	}

	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "post must have content",
		})
		return
	}

	if len(req.Communities) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "post must belong to at least one community",
		})
		return
	}

	communities, err := pr.db.GetCommunities(c, req.Communities)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err,
		})
		log.Fatal("database error occurred", err)
		return
	}
	if len(communities) != len(req.Communities) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "one of the communities does not exist",
		})
		return
	}

	creatorAlias := ""
	if req.Visibility == types.VisibilityHidden {
		creatorAlias = util.GenerateAlias()
	}

	id, err := pr.db.CreatePost(c, &db.CreatePost{
		Title:       req.Title,
		Content:     req.Content,
		Communities: req.Communities,
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:    getUserToken(c).UID,
			Visibility:   req.Visibility,
			CreatorAlias: creatorAlias,
		},
	})
	if err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
	return
}

type createCommentReq struct {
	Content    string           `json:"content"`
	Visibility types.Visibility `json:"visibility"`
}

func (pr *postRoutes) createComment(c *gin.Context) {
	var req createCommentReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "post must have content",
		})
		return
	}

	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		util.HandleHTTPErrorRes(c, httpErr)
		return
	}

	creatorAlias := ""
	if req.Visibility == types.VisibilityHidden {
		creatorAlias = util.GenerateAlias()
	}

	id, err := pr.db.CreateComment(c, &db.CreateComment{
		Content:          req.Content,
		ParentMetadataId: post.ContentMetadata.Id,
		CreateContentMetadata: &db.CreateContentMetadata{
			CreatorId:    getUserToken(c).UID,
			Visibility:   req.Visibility,
			CreatorAlias: creatorAlias,
		},
	})
	if err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
	return
}

func (pr *postRoutes) getPostById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	post, err := pr.db.GetPostById(c, id)
	if err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "post not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    makePostDisplayable(post),
	})
}

func (pr *postRoutes) getPosts(c *gin.Context) {
	var from *time.Time
	if c.Query("from") != "" {
		fromTime, err := time.Parse(time.RFC3339, c.Query("from"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err,
			})
			return
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
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": err,
				})
				return
			}
			communityIds[i] = communityId
		}
	}

	limit := int64(5)
	if c.Query("limit") != "" {
		var err error
		limit, err = strconv.ParseInt(c.Query("limit"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err,
			})
			return
		}
		if limit > 500 {
			limit = 500
		}
	}
	cursor := c.Query("cursor")

	posts, err := pr.db.GetPosts(c, from, cursor, communityIds, int16(limit))
	if err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	displayablePosts := make([]*types.Post, len(posts))
	for i, post := range posts {
		displayablePosts[i] = makePostDisplayable(post)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    displayablePosts,
	})
}

func makePostDisplayable(post *types.Post) *types.Post {
	switch post.Visibility {
	case types.VisibilityHidden:
		post.Creator = &types.DisplayableUser{
			User:  nil,
			Alias: post.Creator.Alias,
		}
	}
	return post
}

type voteReq struct {
	Value int8
}

func (pr *postRoutes) vote(c *gin.Context) {
	postId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	post, err := pr.db.GetPostById(c, postId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		log.Println("a database error occurred", err)
		return
	} else if post == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "post does not exist",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	var req voteReq
	if err := c.BindJSON(&req); err != nil {
		// TODO: Fix validation error messages
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	if req.Value < -1 {
		req.Value = -1
	} else if req.Value > 1 {
		req.Value = 1
	}

	if err := pr.db.Vote(c, getUserToken(c).UID, post.ContentMetadata.Id, req.Value); err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type reportReq struct {
	Reason string
}

// TODO: Add GetReportByPost and GetReportByCommunity
func (pr *postRoutes) report(c *gin.Context) {
	var req reportReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	/* TODO: Possible race condition if post was deleted
	right after report created
	*/
	post, httpErr := pr.mustGetPostByIdStr(c, c.Param("id"))
	if httpErr != nil {
		util.HandleHTTPErrorRes(c, httpErr)
		return
	}

	reportId, err := pr.db.CreateReport(c, getUserToken(c).UID, &db.CreateReport{
		PostId: post.Id,
		Reason: req.Reason,
	})
	if err != nil {
		log.Println("a database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": reportId,
		},
	})
}

// mustGetPostByIdStr attempts to get the post by the id string
func (pr *postRoutes) mustGetPostByIdStr(ctx context.Context, idStr string) (*types.Post, *util.HTTPError) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &util.DbHTTPErr
	}
	post, err := pr.db.GetPostById(ctx, id)
	if err != nil {
		log.Println("a database error occurred", err)
		return nil, &util.MalformedIdHTTPErr
	} else if post == nil {
		return nil, &postDoesNotExistHTTPErr
	}
	return post, nil
}
