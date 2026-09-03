package routes

import (
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/invoices", handlers.InvoicesHandler(pool))
	r.POST("/invoices", handlers.CreateNFSeHandler(pool))
	r.POST("/invoices/:id/print", handlers.PrintInvoiceHandler(pool))

	// Teste de rota para gerar um panic e testar o middleware de recuperação
	r.GET("/test-panic", func(c *gin.Context) {
		panic("panic de teste")
	})
}
