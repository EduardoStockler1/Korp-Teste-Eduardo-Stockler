package routes

import (
	"net/http"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"service": "invoicing",
				"error":   "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "invoicing",
		})
	})

	r.GET("/invoices", handlers.InvoicesHandler(pool))
	r.POST("/invoices", handlers.CreateNFSeHandler(pool))
	r.POST("/invoices/:id/print", handlers.PrintInvoiceHandler(pool))

	// Teste de rota para gerar um panic e testar o middleware de recuperação
	r.GET("/test-panic", func(c *gin.Context) {
		panic("panic de teste")
	})
}
