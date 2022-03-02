package routes

import (
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

func AddPostRoutes(group *gin.RouterGroup, db db.Database) {
	routes := postRoutes{db}
	posts := group.Group("/posts")
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
	// TODO: Add validation and standardize responses
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	createPost := db.CreatePost(req)
	id, err := pr.db.CreatePost(c, &createPost)
	if err != nil {
		log.Println("A database error occurred", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "DB error",
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
	}
	post, err := pr.db.GetPostById(c, id)
	if err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Database error",
		})
		return
	}
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Post not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    makePostDisplayable(post),
	})
}

func makePostDisplayable(post *types.Post) *types.Post {
	if post.Visibility == types.HIDDEN {

	}
	return post
}
