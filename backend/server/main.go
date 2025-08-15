package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	srv := &Server{DB: db}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", srv.registerHandler)
			auth.POST("/login", srv.loginHandler)
			// auth.POST("/logout", srv.logoutHandler) // אופציונלי
			// auth.GET("/validate", srv.validateHandler) // אופציונלי
		}
	}

	log.Println("listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
