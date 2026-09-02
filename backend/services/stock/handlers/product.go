package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/stock/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Cria produto a partir de uma requisição HTTP POST com JSON no corpo
func CreateProductHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		var product models.Product

		// Lê o JSON enviado na requisição
		if err := c.ShouldBindJSON(&product); err != nil {
			log.Error().
				Err(err).
				Str("service", "stock").
				Str("operation", "create_product").
				Msg("JSON inválido")

			c.JSON(http.StatusBadRequest, gin.H{
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
			log.Error().
				Err(err).
				Str("service", "stock").
				Str("operation", "create_product").
				Str("product_code", product.Code).
				Msg("erro ao cadastrar produto")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao cadastrar produto",
			})
			return
		}

		log.Info().
			Str("service", "stock").
			Str("operation", "create_product").
			Int("product_id", product.ID).
			Str("product_code", product.Code).
			Str("description", product.Description).
			Int("stock", product.Stock).
			Msg("produto cadastrado")

		// Retorna o produto criado
		c.JSON(http.StatusCreated, product)
	}
}

// Retorna a lista de produtos em formato JSON
func ProductsHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Consulta os produtos no banco de dados
		rows, err := conn.Query(
			context.Background(),
			`SELECT id, code, description, stock
			 FROM products
			 ORDER BY id`,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "stock").
				Str("operation", "list_products").
				Msg("erro ao buscar produtos")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar produtos",
			})
			return
		}

		defer rows.Close()

		products := []models.Product{}

		// Percorre os produtos retornados pelo banco
		for rows.Next() {
			var product models.Product

			err := rows.Scan(
				&product.ID,
				&product.Code,
				&product.Description,
				&product.Stock,
			)

			if err != nil {
				log.Error().
					Err(err).
					Str("service", "stock").
					Str("operation", "list_products").
					Msg("erro ao ler produto")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao ler produto",
				})
				return
			}

			products = append(products, product)
		}

		log.Info().
			Str("service", "stock").
			Str("operation", "list_products").
			Int("products_count", len(products)).
			Msg("produtos listados")

		// Retorna a lista de produtos
		c.JSON(http.StatusOK, products)
	}
}

// Retorna um produto pelo ID
func ProductHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Warn().
				Err(err).
				Str("service", "stock").
				Str("operation", "get_product").
				Str("product_id", c.Param("id")).
				Msg("ID de produto inválido")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID inválido",
			})
			return
		}

		var product models.Product

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
				log.Warn().
					Str("service", "stock").
					Str("operation", "get_product").
					Int("product_id", id).
					Msg("produto não encontrado")

				c.JSON(http.StatusNotFound, gin.H{
					"error": "Produto não encontrado",
				})
				return
			}

			log.Error().
				Err(err).
				Str("service", "stock").
				Str("operation", "get_product").
				Int("product_id", id).
				Msg("erro ao buscar produto")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar produto",
			})
			return
		}

		log.Info().
			Str("service", "stock").
			Str("operation", "get_product").
			Int("product_id", product.ID).
			Str("product_code", product.Code).
			Msg("produto encontrado")

		c.JSON(http.StatusOK, product)
	}
}
