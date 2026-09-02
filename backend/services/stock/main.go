package main

import (
	"context"
	"os"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/middleware"
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Inicia o servidor HTTP na porta 8081
func main() {

	// Inicializa o logger
	initLogger()

	// Conecta ao banco de dados
	conn, err := connectDatabase()

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "stock").
			Str("operation", "database_connection").
			Msg("erro ao conectar ao banco")
		return
	}

	// Fecha a conexão com o banco quando o programa terminar
	defer conn.Close(context.Background())

	// Testa a conexão com o banco
	err = testDatabaseConnection(conn)

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

	r.Use(middleware.RecoveryMiddleware())

	// Configura as rotas
	routes.SetupRoutes(r, conn)

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
