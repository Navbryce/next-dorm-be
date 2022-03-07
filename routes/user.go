package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/heroku/go-getting-started/db"
	"github.com/heroku/go-getting-started/types"
	"log"
	"net/http"
)

type userRoutes struct {
	db db.UserDatabase
}

func AddUserRoutes(group *gin.RouterGroup, userDatabase db.UserDatabase, authClient *auth.Client) {
	routes := userRoutes{userDatabase}
	users := group.Group("/users", Auth(userDatabase, authClient, &AuthConfig{
		profileNotRequired: true,
	}))
	users.PUT("", routes.CreateUser)
}

type createUserReq struct {
	DisplayName string
}

func (ur userRoutes) CreateUser(c *gin.Context) {
	var req createUserReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"success": false,
			"message": err,
		})
	}
	if err := ur.db.CreateUser(c, &types.User{
		Id:          getUserToken(c).UID,
		DisplayName: req.DisplayName,
	}); err != nil {
		log.Println("database error occurred", err)
		c.JSON(http.StatusInternalServerError, &gin.H{
			"success": false,
			"message": "database error",
		})
		return
	}
	c.JSON(http.StatusOK, &gin.H{
		"success": true,
	})
}
