package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

	// Rotas de produtos
	r.GET("/products/:id", productHandler(conn))
	r.GET("/products", productsHandler(conn))
	r.POST("/products", createProductHandler(conn))

	fmt.Println("Stock Service iniciado em http://localhost:8081")

	// Inicia o servidor
	err = r.Run(":8081")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}

// Cria produto a partir de uma requisição HTTP POST com JSON no corpo
func createProductHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		var product Product

		// Lê o JSON enviado na requisição
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(400, gin.H{
				"error": "JSON inválido",
			})
			return
		}

		// Insere o produto no banco de dados
		err := conn.QueryRow(
			context.Background(),
			`INSERT INTO products (code, description, stock)
			 VALUES ($1, $2, $3)
			 RETURNING id`,
			product.Code,
			product.Description,
			product.Stock,
		).Scan(&product.ID)

		if err != nil {
			c.JSON(500, gin.H{
				"error": "Erro ao cadastrar produto",
			})
			return
		}

		fmt.Printf("Produto cadastrado: %+v\n", product)

		// Retorna o produto criado
		c.JSON(201, product)
	}
}

// Retorna a lista de produtos em formato JSON
func productsHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Consulta os produtos no banco de dados
		rows, err := conn.Query(
			context.Background(),
			`SELECT id, code, description, stock
			 FROM products
			 ORDER BY id`,
		)

		if err != nil {
			c.JSON(500, gin.H{
				"error": "Erro ao buscar produtos",
			})
			return
		}

		defer rows.Close()

		products := []Product{}

		// Percorre os produtos retornados pelo banco
		for rows.Next() {
			var product Product

			err := rows.Scan(
				&product.ID,
				&product.Code,
				&product.Description,
				&product.Stock,
			)

			if err != nil {
				c.JSON(500, gin.H{
					"error": "Erro ao ler produto",
				})
				return
			}

			products = append(products, product)
		}

		// Retorna a lista de produtos
		c.JSON(200, products)
	}
}

// Verifica se um produto existe no banco de dados
func productHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"error": "ID inválido",
			})
			return
		}

		var product Product

		err = conn.QueryRow(
			context.Background(),
			`SELECT id, code, description, stock
			 FROM products
			 WHERE id = $1`,
			id,
		).Scan(
			&product.ID,
			&product.Code,
			&product.Description,
			&product.Stock,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(404, gin.H{
					"error": "Produto não encontrado",
				})
				return
			}

			c.JSON(500, gin.H{
				"error": "Erro ao buscar produto",
			})
			return
		}

		c.JSON(200, product)
	}
}
