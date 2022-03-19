package middleware

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"net/http"
	"strings"
)

const (
	TOKEN_KEY = "authToken"
	USER_KEY  = "user"
)

type AuthConfig struct {
	sessionNotRequired    bool
	appAccountNotRequired bool
}

// TODO: figure out the best way of handling admin only?
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

		c.Set(TOKEN_KEY, token)

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
			if config.appAccountNotRequired {
				return
			}
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "must have a user profile",
			})
			c.Abort()
			return
		}
		c.Set(USER_KEY, user)
	}
}

func GetToken(c *gin.Context) *auth.Token {
	token, _ := c.Get(TOKEN_KEY)
	return token.(*auth.Token)
}

type UserWithToken struct {
	*auth.Token
	*model.User
}

func GetUserWithToken(c *gin.Context) *UserWithToken {
	user, _ := c.Get(USER_KEY)
	return &UserWithToken{
		Token: GetToken(c),
		User:  user.(*model.User),
	}
}
