package routes

import (
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func SetupRoutes(r *gin.Engine, conn *pgx.Conn) {
	r.GET("/products/:id", handlers.ProductHandler(conn))
	r.GET("/products", handlers.ProductsHandler(conn))
	r.POST("/products", handlers.CreateProductHandler(conn))
}
