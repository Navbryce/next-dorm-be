package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"log"
	"net/http"
)

type communityRoutes struct {
	db db.Database
}

func AddCommunityRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := communityRoutes{db}
	posts := group.Group("/communities", Auth(db, authClient, &AuthConfig{}))
	posts.PUT("", routes.createCommunity)
}

type createCommunityReq struct {
	Name string
}

func (cr *communityRoutes) createCommunity(c *gin.Context) {
	var req createCommunityReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	if len(req.Name) <= 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "community name must be more than 5 characters",
		})
	}
	id, err := cr.db.CreateCommunity(c, req.Name)
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
