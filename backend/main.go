package main

import (
	"github.com/Mirac61/Invoice/backend/invoice"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type"},
	}))

	repo := invoice.NewRepository()
	service := invoice.NewService(repo)
	handler := invoice.NewHandler(service)

	r.POST("/api/invoices", handler.Create)
	r.GET("/api/invoices", handler.GetAll)
	r.GET("/api/invoices/:id", handler.GetByID)
	r.DELETE("/api/invoices/:id", handler.Delete)
	r.PUT("/api/invoices/:id", handler.Update)
	//	r.PATCH("/api/invoices/:id", handler.PartialUpdate)

	r.Run(":8080")
}
