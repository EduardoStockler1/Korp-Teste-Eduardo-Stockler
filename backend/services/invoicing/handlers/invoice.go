package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Cria nota fiscal e seus itens no banco de dados
func CreateNFSeHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice models.Invoice

		// Lê e valida o JSON enviado na requisição
		if err := c.ShouldBindJSON(&invoice); err != nil {
			log.Warn().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "create_nfse").
				Msg("Payload inválido")

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Payload inválido",
				"details": err.Error(),
			})
			return
		}

		invoice.Status = "OPEN" //Nota por padrão deve ser gerada com status OPEN, pois ainda não foi impressa

		// Verifica se todos os produtos existem no Stock Service
		for _, item := range invoice.Items {
			err := productExists(item.ProductID)

			if err != nil {
				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "validate_product").
					Int("product_id", item.ProductID).
					Msg("erro ao validar produto no Stock Service")

				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Erro ao buscar produto no stock service",
				})
				return
			}
		}

		// Inicia a transação
		tx, err := conn.Begin(context.Background())
		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "begin_transaction").
				Msg("erro ao iniciar transação")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao iniciar transação",
			})
			return
		}

		// Se algo der errado antes do Commit, desfaz tudo
		defer tx.Rollback(context.Background())

		// Insere a invoice
		err = tx.QueryRow(
			context.Background(),
			`INSERT INTO invoices (number, status)
			 VALUES ($1, $2)
			 RETURNING id`,
			invoice.Number,
			invoice.Status,
		).Scan(&invoice.ID)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "insert_invoice").
				Int("invoice_number", invoice.Number).
				Str("status", invoice.Status).
				Msg("erro ao inserir invoice")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao cadastrar nota fiscal",
			})
			return
		}

		// Insere os itens da invoice
		for i := range invoice.Items {
			item := &invoice.Items[i]

			err = tx.QueryRow(
				context.Background(),
				`INSERT INTO invoice_items (invoice_id, product_id, quantity)
				 VALUES ($1, $2, $3)
				 RETURNING id`,
				invoice.ID,
				item.ProductID,
				item.Quantity,
			).Scan(&item.ID)

			if err != nil {
				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "insert_invoice_item").
					Int("invoice_id", invoice.ID).
					Int("product_id", item.ProductID).
					Int("quantity", item.Quantity).
					Msg("erro ao inserir item da invoice")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao adicionar item à nota fiscal",
				})
				return
			}

			item.InvoiceID = invoice.ID
		}

		// Confirma a transação
		err = tx.Commit(context.Background())
		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "commit_transaction").
				Int("invoice_id", invoice.ID).
				Msg("erro ao comitar transação")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao finalizar transação",
			})
			return
		}

		log.Info().
			Str("service", "invoicing").
			Str("operation", "create_nfse").
			Int("invoice_id", invoice.ID).
			Int("invoice_number", invoice.Number).
			Str("status", invoice.Status).
			Int("items_count", len(invoice.Items)).
			Msg("nota fiscal cadastrada")

		c.Header("Content-Type", "application/json")
		c.Status(http.StatusCreated)

		json.NewEncoder(c.Writer).Encode(invoice)
	}
}

// Retorna a lista de notas fiscais em formato JSON
func InvoicesHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Content-Type", "application/json")

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))

		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "página inválida",
			})
			return
		}

		// Quantidade de registros por página
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Limite inválido",
			})
			return
		}

		if limit > 100 {
			limit = 100
		}

		offset := (page - 1) * limit

		// total de invoices
		var total int

		// Consulta todas as notas fiscais no banco de dados
		err = conn.QueryRow(
			context.Background(),
			`SELECT COUNT(*) FROM invoices`,
		).Scan(&total)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "list_invoices").
				Msg("erro ao buscar notas fiscais")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar notas fiscais",
			})
			return
		}

		rows, err := conn.Query(
			context.Background(),
			`SELECT id, number, status
			 FROM invoices
			 ORDER BY id
		  	 LIMIT $1 OFFSET $2`,
			limit,
			offset,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "list_invoices").
				Msg("erro ao buscar notas fiscais")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar notas fiscais",
			})
			return
		}

		defer rows.Close()

		invoices := []models.Invoice{}

		//
		for rows.Next() {
			var invoice models.Invoice

			err := rows.Scan(
				&invoice.ID,
				&invoice.Number,
				&invoice.Status,
			)

			if err != nil {
				rows.Close()

				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "scan_invoice").
					Msg("erro ao ler nota fiscal")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao ler nota fiscal",
				})
				return
			}

			// Inicializa os itens como uma lista vazia
			invoice.Items = []models.InvoiceItem{}

			// Adiciona a nota à lista
			invoices = append(invoices, invoice)
		}

		// Consulta todos os itens das notas fiscais

		for i := range invoices {

			itemRows, err := conn.Query(
				context.Background(),
				`SELECT id, invoice_id, product_id, quantity
				 FROM invoice_items
				 WHERE invoice_id = $1
				 ORDER BY id`,
				invoices[i].ID,
			)

			if err != nil {
				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "list_invoice_items").
					Int("invoice_id", invoices[i].ID).
					Msg("erro ao buscar itens das notas fiscais")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao buscar itens da nota fiscal",
				})
				return
			}

			for itemRows.Next() {
				var item models.InvoiceItem

				err := itemRows.Scan(
					&item.ID,
					&item.InvoiceID,
					&item.ProductID,
					&item.Quantity,
				)

				if err != nil {
					itemRows.Close()

					log.Error().
						Err(err).
						Str("service", "invoicing").
						Str("operation", "scan_invoice_item").
						Int("invoice_id", item.InvoiceID).
						Msg("erro ao ler item da nota fiscal")

					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Erro ao ler item da nota fiscal",
					})
					return
				}
				invoices[i].Items = append(invoices[i].Items, item)
			}
			itemRows.Close()
		}

		totalPages := (total + limit - 1) / limit

		log.Info().
			Str("service", "invoicing").
			Str("operation", "list_invoices").
			Int("page", page).
			Int("limit", limit).
			Int("total", total).
			Msg("invoices listadas")

		c.JSON(http.StatusOK, gin.H{
			"data":       invoices,
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		})
	}
}

// Verifica se um produto existe no Stock Service
func productExists(productID int) error {
	stockServiceURL := os.Getenv("STOCK_SERVICE_URL")

	if stockServiceURL == "" {
		stockServiceURL = "http://localhost:8081"
	}
	url := fmt.Sprintf(
		"%s/products/%d",
		stockServiceURL,
		productID,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "validate_product").
			Int("product_id", productID).
			Msg("erro ao consultar Stock Service")

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		err := fmt.Errorf("produto %d não encontrado", productID)

		log.Warn().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "validate_product").
			Int("product_id", productID).
			Int("status_code", resp.StatusCode).
			Msg("produto não encontrado no Stock Service")

		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("erro ao consultar Stock Service")

		log.Error().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "validate_product").
			Int("product_id", productID).
			Int("status_code", resp.StatusCode).
			Msg("Stock Service retornou erro")

		return err
	}

	return nil
}
