package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func setupRoutes(r *gin.Engine, conn *pgx.Conn) {
	r.GET("/products/:id", productHandler(conn))
	r.GET("/products", productsHandler(conn))
	r.POST("/products", createProductHandler(conn))
}
