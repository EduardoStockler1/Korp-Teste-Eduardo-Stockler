package routes

import (
	"net/http"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/handlers"

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

	r.GET("/products/:id", handlers.ProductHandler(pool))
	r.GET("/products", handlers.ProductsHandler(pool))
	r.POST("/products", handlers.CreateProductHandler(pool))
	r.POST("/products/:id/decrease-stock", handlers.DecreaseStockHandler(pool))

	// Teste de rota para gerar um panic e testar o middleware de recuperação
	r.GET("/test-panic", func(c *gin.Context) {
		panic("panic de teste")
	})
}
