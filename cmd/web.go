package main

import (
	"context"
	"github.com/heroku/go-getting-started/db"
	"github.com/heroku/go-getting-started/routes"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := db.GetDatabase()
	if err != nil {
		log.Fatal("Received err when attempting to connect to DB", err)
	}
	defer db.Close()

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing firebase: %v\n", err)
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatal("error initializing auth client", err)
	}

	// TODO: Move the port parsing to a configuration module
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	routes.AddCommunityRoutes(&r.RouterGroup, db, authClient)
	routes.AddPostRoutes(&r.RouterGroup, db, authClient)
	routes.AddUserRoutes(&r.RouterGroup, db, authClient)

	if err := r.Run(); err != nil {
		log.Fatal("Error when attempting to run web server", err)
	}
}
