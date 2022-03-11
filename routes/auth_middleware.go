package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/heroku/go-getting-started/db"
	"net/http"
	"strings"
)

const (
	AUTH_TOKEN_KEY   = "authToken"
	USER_PROFILE_KEY = "user"
)

type AuthConfig struct {
	sessionNotRequired bool
	profileNotRequired bool
}

func Auth(userDB db.UserDatabase, authClient *auth.Client, config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader, ok := c.Request.Header["Authorization"]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "no authorization header",
			})
			c.Abort()
			return
		}
		if len(authorizationHeader) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "no authorization header",
			})
			c.Abort()
			return
		}
		if strings.Index(authorizationHeader[0], "Bearer ") != 0 || len(authorizationHeader[0]) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "incorrectly formatted authorization header",
			})
			c.Abort()
			return
		}
		token, err := authClient.VerifyIDToken(c, authorizationHeader[0][7:])

		c.Set(AUTH_TOKEN_KEY, token)

		if err != nil {
			if config.sessionNotRequired {
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid token",
			})
			c.Abort()
			return
		}

		user, err := userDB.GetUser(c, token.UID)
		if user == nil {
			if config.profileNotRequired {
				return
			}
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "must have a user profile",
			})
			c.Abort()
			return
		}
		c.Set(USER_PROFILE_KEY, user)
	}
}

func getUserToken(c *gin.Context) *auth.Token {
	token, _ := c.Get(AUTH_TOKEN_KEY)
	return token.(*auth.Token)
}
