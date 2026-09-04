package main

import (
	"os"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/middleware"
	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {

	// conecta ao banco de dados
	pool, err := connectDatabase()

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao conectar ao banco")
		return
	}

	// fecha o pool de conexões quando o programa terminar
	defer pool.Close()

	// testa a conexão com o banco
	err = testDatabaseConnection(pool)

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao testar conexão")
		return
	}

	r := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.Use(middleware.RecoveryMiddleware())

	routes.SetupRoutes(r, pool)

	log.Info().
		Str("service", "invoicing").
		Str("address", "Invoicing Service iniciado em http://localhost:8082").
		Msg("Invoicing Service iniciado")

	// inicia o servidor
	port := os.Getenv("INVOICING_PORT")

	if port == "" {
		port = "8082"
	}

	err = r.Run(":" + port)

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao iniciar servidor")
	}
}
