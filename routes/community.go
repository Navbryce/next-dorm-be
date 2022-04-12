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
	posts.GET("", util.HandlerWrapper(routes.getCommunities, &util.HandlerOpts{}))
	posts.GET("/:id", util.HandlerWrapper(routes.getCommunityById, &util.HandlerOpts{}))
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
	communities, err := cr.db.GetCommunities(c, []int64{id}, &db.GetCommunitiesQueryOpts{
		ForUserId: middleware.GetUserIdMaybe(c),
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	if len(communities) == 0 {
		return nil, nil
	}
	return communities[0], nil
}

func (cr *communityRoutes) getCommunities(c *gin.Context) (interface{}, *util.HTTPError) {
	communities, err := cr.db.GetCommunities(c, nil, &db.GetCommunitiesQueryOpts{
		ForUserId: middleware.GetUserIdMaybe(c),
	})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return communities, nil
}
