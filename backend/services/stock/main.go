package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Inicia o servidor HTTP na porta 8081
func main() {

	// Conecta ao banco de dados
	conn, err := connectDatabase()

	if err != nil {
		fmt.Println("Erro ao conectar ao banco:", err)
		return
	}

	// Fecha a conexão com o banco quando o programa terminar
	defer conn.Close(context.Background())

	// Testa a conexão com o banco
	err = testDatabaseConnection(conn)

	if err != nil {
		fmt.Println("Erro ao testar conexão:", err)
		return
	}

	// Cria o servidor Gin
	r := gin.Default()

	routes.setupRoutes(r, conn)

	fmt.Println("Stock Service iniciado em http://localhost:8081")

	// Inicia o servidor
	err = r.Run(":8081")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}
