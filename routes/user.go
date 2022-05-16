package routes

import (
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/services"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
	"strings"
)

const MinDisplayNameLength = 4

type userRoutes struct {
	db         db.UserDatabase
	userBucket *services.StorageBucket
}

func AddUserRoutes(group *gin.RouterGroup, userDatabase db.UserDatabase, authClient *auth.Client, userBucket *services.StorageBucket) {
	routes := userRoutes{userDatabase, userBucket}
	users := group.Group("/users", middleware.GenAuth(userDatabase, authClient, &middleware.AuthConfig{}))
	users.GET("/:userId", util.HandlerWrapper(routes.GetLocalUser, &util.HandlerOpts{}))
	users.PUT("",
		middleware.RequireToken(),
		util.HandlerWrapper(routes.CreateLocalUser, &util.HandlerOpts{}))
	users.GET("",
		middleware.RequireToken(),
		util.HandlerWrapper(routes.GetCurrentLocalUser, &util.HandlerOpts{}))
}

type createUserReq struct {
	DisplayName string `json:"displayName"`
}

func (cur *createUserReq) Sanitize() *createUserReq {
	return &createUserReq{DisplayName: util.XSSSanitize(cur.DisplayName)}
}

func (cur *createUserReq) Validate() *util.HTTPError {
	if len(cur.DisplayName) < MinDisplayNameLength {
		return &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("display name must be larger than %v", MinDisplayNameLength),
		}
	}
	return nil
}

func (ur userRoutes) CreateLocalUser(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createUserReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	req = *req.Sanitize()

	if err := req.Validate(); err != nil {
		return nil, err
	}
	if exists, err := ur.userBucket.Exists(c, middleware.MustGetToken(c).AvatarBlobNameForUser()); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	} else if !exists {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "must have an avatar",
		}
	}

	user := &model.LocalUser{
		Id:          middleware.MustGetToken(c).UID,
		DisplayName: req.DisplayName,
	}
	if err := ur.db.CreateUser(c, user); err != nil {
		err, ok := err.(*mysql.MySQLError)
		if ok && db.IsDupKeyErr(err) {
			if strings.Contains(db.GetDupKey(err), "display_name") {
				return nil, &util.HTTPError{
					Status:  http.StatusConflict,
					Message: "display name is taken",
				}
			}
			return nil, &util.HTTPError{
				Status:  http.StatusConflict,
				Message: "profile already exists",
			}
		}
		return nil, util.BuildDbHTTPErr(err)
	}
	// return the user to get the generated avatar if not specified in req
	return user, nil
}
func (ur userRoutes) GetCurrentLocalUser(c *gin.Context) (interface{}, *util.HTTPError) {

	user, err := ur.db.GetUser(c, middleware.MustGetToken(c).UID)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return user, nil
}

func (ur userRoutes) GetLocalUser(c *gin.Context) (interface{}, *util.HTTPError) {
	userId := c.Param("userId")
	user, err := ur.db.GetUser(c, userId)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	if user == nil {
		return nil, nil
	}
	// TODO: Validate security of this endpoint
	return user.MakeDisplayableFor(middleware.GetLocalUser(c)), nil
}
