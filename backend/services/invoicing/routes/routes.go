package routes

import (
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func SetupRoutes(r *gin.Engine, conn *pgx.Conn) {
	r.GET("/invoices", handlers.InvoicesHandler(conn))
	r.POST("/invoices", handlers.CreateNFSeHandler(conn))
}
