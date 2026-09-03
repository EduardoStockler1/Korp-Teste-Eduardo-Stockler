package main

import (
	"os"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/middleware"
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Inicia o servidor HTTP na porta 8081
func main() {

	// Inicializa o logger
	initLogger()

	// Conecta ao banco de dados
	pool, err := connectDatabase()

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "stock").
			Str("operation", "database_connection").
			Msg("erro ao conectar ao banco")
		return
	}

	// Fecha o pool de conexões quando o programa terminar
	defer pool.Close()

	// Testa a conexão com o banco
	err = testDatabaseConnection(pool)

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "stock").
			Str("operation", "database_ping").
			Msg("erro ao testar conexão com banco")
		return
	}

	// Cria o servidor Gin
	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.Use(middleware.RecoveryMiddleware())

	// Configura as rotas
	routes.SetupRoutes(r, pool)

	log.Info().
		Str("service", "stock").
		Str("address", "http://localhost:8081").
		Msg("Stock Service iniciado")

	// Inicia o servidor
	port := os.Getenv("STOCK_PORT")

	if port == "" {
		port = "8081"
	}

	err = r.Run(":" + port)

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "stock").
			Str("operation", "server_start").
			Msg("erro ao iniciar servidor")
	}
}
