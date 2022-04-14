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
}

// TODO: figure out the best way of handling admin only?
func GenAuth(userDB db.UserDatabase, authClient *auth.Client, _ *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader, ok := c.Request.Header["Authorization"]
		if !ok {
			return
		}
		if len(authorizationHeader) == 0 {
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
		// TODO: VerifyIDToken and check revoked?
		token, err := authClient.VerifyIDToken(c, authorizationHeader[0][7:])

		c.Set(TOKEN_KEY, token)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid token",
			})
			c.Abort()
			return
		}

		// TODO: Make hasAccount a custom claim on ID token to short-circuit DB query
		user, err := userDB.GetUser(c, token.UID)
		if user == nil {
			return
		}
		c.Set(USER_KEY, user)
	}
}

func RequireToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetToken(c) != nil {
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid token",
		})
		c.Abort()
	}
}

// RequireAccount requires an account (and a token)
func RequireAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		RequireToken()(c)
		if c.IsAborted() {
			return
		}
		if GetUser(c) != nil {
			return
		}
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "must have a user profile",
		})
		c.Abort()
		return
	}
}

func GetToken(c *gin.Context) *auth.Token {
	tokenMaybe, loggedIn := c.Get(TOKEN_KEY)
	if !loggedIn {
		return nil
	}
	return tokenMaybe.(*auth.Token)
}

func MustGetToken(c *gin.Context) *auth.Token {
	token := GetToken(c)
	if token == nil {
		panic("expected a token")
	}
	return token
}

func GetUser(c *gin.Context) *model.User {
	userMaybe, hasProfile := c.Get(USER_KEY)
	if !hasProfile {
		return nil
	}
	return userMaybe.(*model.User)
}

func MustGetUser(c *gin.Context) *model.User {
	user := GetUser(c)
	if user == nil {
		panic("expected a user to be logged in")
	}
	return user
}

func GetUserIdMaybe(c *gin.Context) (userId string) {
	if token := GetToken(c); token != nil {
		userId = token.UID
	}
	return userId
}
