package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
)

type communityRoutes struct {
	db db.Database
}

func AddCommunityRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := communityRoutes{db}
	posts := group.Group("/communities", middleware.GenAuth(db, authClient, &middleware.AuthConfig{}))
	posts.PUT("", util.HandlerWrapper(routes.createCommunity, &util.HandlerOpts{}))
}

type createCommunityReq struct {
	Name string
}

func (cr *communityRoutes) createCommunity(c *gin.Context) (interface{}, *util.HTTPError) {
	var req createCommunityReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}
	if len(req.Name) <= 5 {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: "community name must be more than 5 characters",
		}
	}
	id, err := cr.db.CreateCommunity(c, req.Name)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return gin.H{
		"id": id,
	}, nil
}
