package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/util"
)

func AddHealthCheckRoutes(group *gin.RouterGroup) {
	health := group.Group("/health")
	health.GET("", util.HandlerWrapper(AliveCheck, &util.HandlerOpts{}))
}

func AliveCheck(c *gin.Context) (interface{}, *util.HTTPError) {
	return nil, nil
}
