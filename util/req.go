package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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
	MalformedIdHTTPErr = &HTTPError{
		Message: "id malformed",
		Status:  http.StatusBadRequest,
	}
)

func BuildDoesNotExistHTTPErr(entityName string) *HTTPError {
	return &HTTPError{
		Message: fmt.Sprintf("%v does not exist", entityName),
		Status:  http.StatusNotFound,
	}
}

func BuildDbHTTPErr(error error) *HTTPError {
	log.Println("database error occurred", error)
	return &HTTPError{
		Status:  http.StatusInternalServerError,
		Message: "database error",
	}
}

func BuildJSONBindHTTPErr(err error) *HTTPError {
	return &HTTPError{
		Message: err.Error(),
		Status:  http.StatusBadRequest,
	}
}

func BuildOperationForbidden(reason string) *HTTPError {
	return &HTTPError{
		Message: fmt.Sprintf("Operation forbidden: %v", reason),
		Status:  http.StatusForbidden,
	}
}

type Handler = func(c *gin.Context) (interface{}, *HTTPError)
type HandlerOpts struct {
}

// HandlerWrapper wraps handlers to provide standardized responses for the API
func HandlerWrapper(route Handler, opts *HandlerOpts) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := route(c)
		if err != nil {
			HandleHTTPErrorRes(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    res,
		})

	}
}

/*
	HandleHTTPErrorRes handles creating the appropriate response for the HTTP error.
	break the route after calling this function

	TODO: Transition to the HandlerWrapper and remove this function
*/
func HandleHTTPErrorRes(c *gin.Context, err *HTTPError) {
	c.JSON(err.Status, gin.H{
		"success": false,
		"message": err.Message,
	})
	return
}
