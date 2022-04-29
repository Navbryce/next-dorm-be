package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/controllers"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
)

type communityRoutes struct {
	db         db.Database
	controller *controllers.CommunityController
}

func AddCommunityRoutes(group *gin.RouterGroup, db db.Database, controller *controllers.CommunityController, authClient *auth.Client) {
	routes := communityRoutes{db, controller}
	posts := group.Group("/communities", middleware.GenAuth(db, authClient, &middleware.AuthConfig{}))
	posts.GET("/:id", util.HandlerWrapper(routes.getCommunityById, &util.HandlerOpts{}))
	posts.GET("/:id/pos", util.HandlerWrapper(routes.getCommunityPos, &util.HandlerOpts{}))
	//posts.PUT("", util.HandlerWrapper(routes.createCommunity, &util.HandlerOpts{}))
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

func (cr *communityRoutes) getCommunityById(c *gin.Context) (interface{}, *util.HTTPError) {
	id, httpErr := util.ParseId(c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	return cr.controller.GetCommunityById(c, id, &db.GetCommunitiesQueryOpts{
		ForUserId: middleware.GetUserIdMaybe(c),
	})
}

func (cr *communityRoutes) getCommunityPos(c *gin.Context) (interface{}, *util.HTTPError) {
	id, httpErr := util.ParseId(c.Param("id"))
	if httpErr != nil {
		return nil, httpErr
	}
	communityPos, err := cr.controller.GetCommunityPos(c, id)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return communityPos, nil
}
