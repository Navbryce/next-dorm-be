package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/heroku/go-getting-started/db"
	"github.com/heroku/go-getting-started/types"
	"log"
	"net/http"
	"strconv"
)

type postRoutes struct {
	db db.Database
}

func AddPostRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := postRoutes{db}
	posts := group.Group("/posts", Auth(db, authClient, &AuthConfig{}))
	posts.GET("", routes.getPosts)
	posts.GET("/:id", routes.getPostById)
	posts.PUT("", routes.createPost)
}

type createPostReq struct {
	Content     string
	Communities []int64
	Visibility  types.Visibility
}

func (pr *postRoutes) createPost(c *gin.Context) {
	var req createPostReq
	// TODO: Add validation and standardize responses. Validate community exists
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
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
	// TODO: Add in req params
	posts, err := pr.db.GetPosts(c, nil, "", []int64{1, 2}, 5)
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
	if post.Visibility == types.HIDDEN {

	}
	return post
}
