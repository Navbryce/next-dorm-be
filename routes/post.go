package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/heroku/go-getting-started/db"
	"github.com/heroku/go-getting-started/types"
	"log"
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
	posts := group.Group("/posts", Auth(db, authClient, &AuthConfig{}))
	posts.GET("", routes.getPosts)
	posts.PUT("", routes.createPost)
	posts.GET("/:id", routes.getPostById)
	posts.PUT("/:id/vote", routes.vote)
}

type createPostReq struct {
	Content     string
	Communities []int64
	Visibility  types.Visibility
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

	createPost := db.CreatePost(req)
	id, err := pr.db.CreatePost(c, getUserToken(c).UID, &createPost)
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
		post.Creator = nil
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
	if err := pr.db.Vote(c, getUserToken(c).UID, postId, req.Value); err != nil {
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
