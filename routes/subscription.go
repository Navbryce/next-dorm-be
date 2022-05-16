package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/middleware"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"net/http"
)

type subscriptionRoutes struct {
	db db.Database
}

func AddSubscriptionRoutes(group *gin.RouterGroup, db db.Database, authClient *auth.Client) {
	routes := subscriptionRoutes{db: db}
	subs := group.Group("/subscriptions", middleware.GenAuth(db, authClient, &middleware.AuthConfig{}))
	subs.POST("", util.HandlerWrapper(routes.subscribe, &util.HandlerOpts{}))
	subs.GET("", util.HandlerWrapper(routes.getSubscriptions, &util.HandlerOpts{}))
}

type subscribeReq = map[int64]bool // subscribeReq represents community ID's to create/delete subscription
// subscribe will stop the moment any individual subscription action fails. partial operations are possible
func (sr *subscriptionRoutes) subscribe(c *gin.Context) (interface{}, *util.HTTPError) {
	var req subscribeReq
	if err := c.BindJSON(&req); err != nil {
		return nil, util.BuildJSONBindHTTPErr(err)
	}

	communityIds := make([]int64, len(req))
	i := 0
	for communityId, _ := range req {
		communityIds[i] = communityId
		i++
	}

	// TODO: susceptible to a race condition
	if fetchedCommunities, err := sr.db.GetCommunitiesByIds(c, communityIds, &db.GetCommunitiesQueryOpts{}); err != nil {
		return nil, util.BuildDbHTTPErr(err)
	} else if len(fetchedCommunities) != len(communityIds) {
		return nil, &util.HTTPError{Status: http.StatusBadRequest, Message: "at least one of the communities does not exist"}
	}

	for communityId, subscribed := range req {
		subAction := sr.db.CreateSubForUser
		if !subscribed {
			subAction = sr.db.DeleteSubForUser
		}
		if err := subAction(c, &model.Subscription{
			UserId:      middleware.MustGetLocalUser(c).Id,
			CommunityId: communityId,
		}); err != nil {
			err, ok := err.(*mysql.MySQLError)
			if !ok || !db.IsDupKeyErr(err) {
				return nil, util.BuildDbHTTPErr(err)
			}
		}
	}
	return nil, nil
}

func (sr *subscriptionRoutes) getSubscriptions(c *gin.Context) (interface{}, *util.HTTPError) {
	subs, err := sr.db.GetSubsForUser(c, middleware.MustGetToken(c).UID)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return subs, nil
}
