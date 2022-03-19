package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"log"
	"net/http"
)

type userRoutes struct {
	db db.UserDatabase
}

func AddUserRoutes(group *gin.RouterGroup, userDatabase db.UserDatabase, authClient *auth.Client) {
	routes := userRoutes{userDatabase}
	users := group.Group("/users", middleware.Auth(userDatabase, authClient, &middleware.AuthConfig{
		appAccountNotRequired: true,
	}))
	users.PUT("", routes.CreateUser)
}

type createUserReq struct {
	DisplayName string
}

// TODO User joins communities?

func (ur userRoutes) CreateUser(c *gin.Context) {
	var req createUserReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"success": false,
			"message": err,
		})
	}
	if err := ur.db.CreateUser(c, &model.User{
		Id:          middleware.GetToken(c).UID,
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
