package handlers

import (
	"log"
	"net/http"
	"product-api/internal/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp" // RabbitMQ go get github.com/streadway/amqp
)

type ProductHandler struct {
	DB         *db.DB
	RabbitConn *amqp.Connection
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product db.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log product data for debugging
	log.Printf("Received product: %+v", product)

	// Store the product in the database
	if err := h.DB.CreateProduct(&product); err != nil {
		log.Printf("Error creating product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Push image URLs to RabbitMQ queue for processing
	err := h.pushToImageQueue(product.Images)
	if err != nil {
		log.Printf("Error pushing to image queue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue images"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) pushToImageQueue(images []string) error {
	// Establish a channel to RabbitMQ
	channel, err := h.RabbitConn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	// Declare a queue for image processing
	queue, err := channel.QueueDeclare(
		"image_processing_queue", // queue name
		true,                     // durable
		false,                    // auto-delete
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return err
	}

	// Push the image URLs to the queue
	for _, image := range images {
		err := channel.Publish(
			"",         // exchange
			queue.Name, // routing key (queue name)
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(image), // Image URL as message
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.DB.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	userID := c.Query("user_id")
	priceMin := c.Query("price_min")
	priceMax := c.Query("price_max")
	name := c.Query("name")

	filters := db.ProductFilters{
		UserID:   userID,
		PriceMin: priceMin,
		PriceMax: priceMax,
		Name:     name,
	}

	products, err := h.DB.ListProducts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, products)
}
