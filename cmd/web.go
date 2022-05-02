package main

import (
	"context"
	"github.com/navbryce/next-dorm-be/controllers"
	"github.com/navbryce/next-dorm-be/db/planetscale"
	"github.com/navbryce/next-dorm-be/routes"
	"github.com/navbryce/next-dorm-be/services"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := planetscale.GetDatabase()
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

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Origin", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	userBucket, err := services.NewStorageBucket(context.Background(), app, "next-dorm-d5c03.appspot.com")
	if err != nil {
		log.Fatal("An error occurred while connecting to the user uploads bucket", err)
	}

	communityController, err := controllers.NewCommunityController(context.Background(), db)
	if err != nil {
		log.Fatal("An error occurred while initializing the community controller", err)
	}

	routes.AddCommunityRoutes(&r.RouterGroup, db, communityController, authClient)
	routes.AddPostRoutes(&r.RouterGroup, db, authClient, userBucket)
	routes.AddSubscriptionRoutes(&r.RouterGroup, db, authClient)
	routes.AddUserRoutes(&r.RouterGroup, db, authClient)

	if err := r.Run(); err != nil {
		log.Fatal("Error when attempting to run web server", err)
	}
}
