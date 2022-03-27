package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/app"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
)

type feedRoutes struct {
	db db.Database
}

func AddFeedRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := feedRoutes{db: db}
	feeds := group.Group("/feeds", middleware.GenAuth(db, authClient, &middleware.AuthConfig{}))
	feeds.POST("/", util.HandlerWrapper(routes.getFeed, &util.HandlerOpts{}))
}

type getFeedReq struct {
	Type   app.PostCursorType
	Cursor map[string]interface{}
}

func (fr *feedRoutes) getFeed(c *gin.Context) (interface{}, *util.HTTPError) {
	var req getFeedReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}
	page, err := app.GetFeedForUser(c, fr.db, middleware.MustGetUser(c), req.Type, req.Cursor)
	if err != nil {
		return nil, &util.HTTPError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}
	posts, cursor, err := page.Posts(c, &app.PostCursorOpts{Limit: 20})
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}

	return &gin.H{
		"posts":  posts,
		"cursor": cursor,
	}, nil
}
