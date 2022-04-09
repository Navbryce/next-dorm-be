package routes

import (
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
)

type userRoutes struct {
	db db.UserDatabase
}

func AddUserRoutes(group *gin.RouterGroup, userDatabase db.UserDatabase, authClient *auth.Client) {
	routes := userRoutes{userDatabase}
	users := group.Group("/users", middleware.GenAuth(userDatabase, authClient, &middleware.AuthConfig{}), middleware.RequireToken())
	users.PUT("", util.HandlerWrapper(routes.CreateUser, &util.HandlerOpts{}))
	users.GET("", util.HandlerWrapper(routes.GetUser, &util.HandlerOpts{}))
}

type createUserReq struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

const MinDisplayNameLength = 4

func (ur userRoutes) CreateUser(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createUserReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"success": false,
			"message": err,
		})
	}
	if len(req.DisplayName) < MinDisplayNameLength {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("display name must be larger than %v", MinDisplayNameLength),
		}
	}
	if len(req.Avatar) == 0 {
		req.Avatar = util.Avatar(req.DisplayName)
	}
	user := &model.User{
		Id:          middleware.MustGetToken(c).UID,
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
	}
	if err := ur.db.CreateUser(c, user); err != nil {
		if db.IsDupKeyErr(err) {
			return nil, &util.HTTPError{
				Status:  http.StatusBadRequest,
				Message: "profile already exists",
			}
		}
		return nil, util.BuildDbHTTPErr(err)
	}
	// return the user to get the generated avatar if not specified in req
	return user, nil
}

func (ur userRoutes) GetUser(c *gin.Context) (interface{}, *util.HTTPError) {
	user, err := ur.db.GetUser(c, middleware.MustGetToken(c).UID)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return user, nil
}
