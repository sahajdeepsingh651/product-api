package main

import (
	"log"
	"product-api/handlers"
	"product-api/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database connection
	database, err := db.NewDB("postgres://postgres:sahaj@localhost:5432/product_management?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Create product handler (update this line to match the new constructor)
	productHandler := handlers.ProductHandler{DB: database}

	// ProductHandler:
	// Define routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the Product API",
		})
	})
	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products/:id", productHandler.GetProductByID)
	router.GET("/products", productHandler.ListProducts)

	// Start the server
	router.Run(":8080")
}
