package main

import (
	"equipment-borrow-system/config"
	"equipment-borrow-system/routes"
	"equipment-borrow-system/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()
	log.Println("Database initialized successfully")

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			utils.Success(c, nil)
			c.Abort()
			return
		}
		c.Next()
	})

	routes.SetupRoutes(r)

	r.NoRoute(func(c *gin.Context) {
		utils.NotFound(c, "API endpoint not found")
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
