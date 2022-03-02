package main

import (
	"github.com/heroku/go-getting-started/db"
	"github.com/heroku/go-getting-started/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	db, err := db.GetDatabase()
	if err != nil {
		log.Fatal("Received err when attempting to connect to DB", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	routes.AddCommunityRoutes(&r.RouterGroup, db)
	routes.AddPostRoutes(&r.RouterGroup, db)

	if err := r.Run(); err != nil {
		log.Fatal("Error when attempting to run web server", err)
	}
}
