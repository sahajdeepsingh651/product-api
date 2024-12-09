package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type Product struct {
	ID               int      `json:"id"`
	UserID           int      `json:"user_id"`
	Name             string   `json:"product_name"`
	Description      string   `json:"product_description"`
	Images           []string `json:"product_images"`
	CompressedImages []string `json:"compressed_product_images"`
	Price            float64  `json:"product_price"`
}

type ProductFilters struct {
	UserID   string
	PriceMin string
	PriceMax string
	Name     string
}

type DB struct {
	conn *sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	err = conn.Ping()
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (db *DB) CreateProduct(product *Product) error {
	query := `
		SELECT id, user_id, product_name, product_description, product_images, compressed_product_images, product_price
		FROM products
		WHERE user_id = $1
	`
	images := "{'" + join(product.Images, "','") + "'}"
	compressedImages := "{}" // Initially empty
	err := db.conn.QueryRow(query, product.UserID, product.Name, product.Description, images, compressedImages, product.Price).Scan(&product.ID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return fmt.Errorf("failed to insert product: %w", err)
	}
	return nil
}

func (db *DB) GetProductByID(id int) (*Product, error) {
	// Query to retrieve product details by ID
	query := `
		SELECT id, user_id, product_name, product_description, product_images, compressed_product_images, product_price
		FROM products WHERE id = $1
	`

	// Execute the query to fetch the product details
	row := db.conn.QueryRow(query, id)

	var product Product
	var images, compressedImages string

	// Scan the results into the product struct
	err := row.Scan(&product.ID, &product.UserID, &product.Name, &product.Description, &images, &compressedImages, &product.Price)
	if err != nil {
		// Return an error if the product is not found or any other issue
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Process and split the image data into separate image URLs
	product.Images = split(images)
	product.CompressedImages = split(compressedImages)

	// If additional image processing is required, you can implement it here.
	// Example: Apply image transformation or processing (like resizing or adding a watermark).
	// For example, if you wanted to resize the images, you could call a function like this:
	// product.Images = processImages(product.Images)

	// Return the populated product object
	return &product, nil
}

func (db *DB) ListProducts(filters ProductFilters) ([]Product, error) {
	query := `
		SELECT id, user_id, product_name, product_description, product_images, compressed_product_images, product_price
		FROM products
		WHERE user_id = $1
	`
	args := []interface{}{filters.UserID}

	// Optional price range filter
	if filters.PriceMin != "" {
		query += " AND product_price >= $2"
		args = append(args, filters.PriceMin)
	}
	if filters.PriceMax != "" {
		query += " AND product_price <= $3"
		args = append(args, filters.PriceMax)
	}

	// Optional product name filter
	if filters.Name != "" {
		query += " AND product_name ILIKE $4"
		args = append(args, "%"+filters.Name+"%")
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var images, compressedImages string
		err := rows.Scan(&product.ID, &product.UserID, &product.Name, &product.Description, &images, &compressedImages, &product.Price)
		if err != nil {
			return nil, err
		}

		product.Images = split(images)
		product.CompressedImages = split(compressedImages)
		products = append(products, product)
	}

	return products, nil
}

func join(arr []string, sep string) string {
	result := ""
	for i, v := range arr {
		if i > 0 {
			result += sep
		}
		result += v
	}
	return result
}

func split(s string) []string {
	return []string{} // Implement splitting logic based on PostgreSQL array format
}
func (db *DB) UpdateCompressedImages(productID int, compressedImages []string) error {
	query := `UPDATE products SET compressed_product_images = $1 WHERE id = $2`
	images := "{'" + join(compressedImages, "','") + "'}"
	_, err := db.conn.Exec(query, images, productID)
	return err
}
