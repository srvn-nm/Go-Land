package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"path/filepath"
	"time"
)

// Basket model
type Basket struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      string    `json:"data" gorm:"size:2048"`
	State     string    `json:"state" gorm:"size:10"`
	UserID    uint      `json:"user_id"`
}

// User model
type User struct {
	ID       uint `gorm:"primary_key ; autoIncrement"`
	Username string
	Password string
}

var db *gorm.DB

func main() {
	// Initialize the Gin router
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	// Connect to the SQLite database
	var err error
	absPath, _ := filepath.Abs("test.db")
	db, err = gorm.Open("sqlite3", absPath)
	if err != nil {
		fmt.Println("Failed to connect to database")
		fmt.Println(err)
		return
	}
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(db)

	// AutoMigrate the models
	db.AutoMigrate(&Basket{}, &User{}) // Add a User model

	// Define API routes
	router.GET("/basket", getBaskets)
	router.POST("/basket", createBasket)
	router.PATCH("/basket/:id", updateBasket)
	router.GET("/basket/:id", getBasketByID)
	router.DELETE("/basket/:id", deleteBasket)

	// Run the server
	err2 := router.Run(":8080")
	if err2 != nil {
		fmt.Println(err2)
	}
}

// Handlers

func getBaskets(c *gin.Context) {
	var baskets []Basket
	db.Find(&baskets)
	c.JSON(200, baskets)
}

func createBasket(c *gin.Context) {
	var basket Basket
	err := c.BindJSON(&basket)
	if err != nil {
		fmt.Println(err)
	}

	// Set default values
	basket.CreatedAt = time.Now()
	basket.UpdatedAt = time.Now()
	basket.State = "PENDING"

	// Save to the database
	db.Create(&basket)

	c.JSON(200, basket)
}

func updateBasket(c *gin.Context) {
	id := c.Param("id")
	var basket Basket
	if err := db.Where("id = ?", id).First(&basket).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	var updatedBasket Basket
	err := c.BindJSON(&updatedBasket)
	if err != nil {
		fmt.Println(err)
	}

	// Update only allowed fields
	basket.Data = updatedBasket.Data
	basket.State = updatedBasket.State
	basket.UpdatedAt = time.Now()

	// Check if the basket is completed
	if basket.State == "COMPLETED" {
		c.JSON(400, gin.H{"error": "Cannot update a completed basket"})
		return
	}

	// Save the updated basket to the database
	db.Save(&basket)

	c.JSON(200, basket)
}

func getBasketByID(c *gin.Context) {
	id := c.Param("id")
	var basket Basket
	if err := db.Where("id = ?", id).First(&basket).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}
	c.JSON(200, basket)
}

func deleteBasket(c *gin.Context) {
	id := c.Param("id")
	var basket Basket
	if err := db.Where("id = ?", id).First(&basket).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	// Delete the basket from the database
	db.Delete(&basket)

	c.JSON(200, gin.H{"message": "Basket deleted"})
}
