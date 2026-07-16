package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type"},
	}))

	r.GET("/api/invoices", getInvoices)
	r.POST("/api/invoices", createInvoice)

	r.Run(":8080")
}

func getInvoices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"invoices": []string{}})
}

func createInvoice(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}
