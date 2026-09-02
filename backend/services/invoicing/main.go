package main

import (
	"context" //banco de dados
	// json para requisições HTTP
	// biblioteca padrão para formatação de strings

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/routes"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	// biblioteca para criar servidor HTTP com rotas usando gin framework
	// biblioteca para conectar ao PostgreSQL
)

func main() {

	// conecta ao banco de dados
	conn, err := connectDatabase()

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao conectar ao banco")
		return
	}

	// fecha a conexão com o banco quando o programa terminar
	defer conn.Close(context.Background())

	// testa a conexão com o banco
	err = testDatabaseConnection(conn)

	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao testar conexão")
		return
	}

	r := gin.Default()
	routes.SetupRoutes(r, conn)

	log.Info().
		Str("service", "invoicing").
		Str("address", "Invoicing Service iniciado em http://localhost:8082").
		Msg("Invoicing Service iniciado")

	// inicia o servidor
	err = r.Run(":8082")
	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Msg("Erro ao iniciar servidor")
	}
}
