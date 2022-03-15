package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HTTPError struct {
	Status  int
	Message string
}

func (he *HTTPError) Error() string {
	return fmt.Sprintf("%v (statusCode=%v)", he.Message, he.Status)
}

var (
	DbHTTPErr = HTTPError{
		Message: "database error",
		Status:  http.StatusInternalServerError,
	}
	MalformedIdHTTPErr = HTTPError{
		Message: "id malformed",
		Status:  http.StatusBadRequest,
	}
)

/*
	HandleHTTPErrorRes handles creating the appropriate response for the HTTP error.
	break the route after calling this function
*/
func HandleHTTPErrorRes(c *gin.Context, err *HTTPError) {
	c.JSON(err.Status, gin.H{
		"success": false,
		"message": err.Message,
	})
	return
}
