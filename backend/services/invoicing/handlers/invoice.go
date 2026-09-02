package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "create_nfse").
				Msg("JSON inválido")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "JSON inválido",
			})
			return
		}

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

		// Consulta todas as notas fiscais no banco de dados
		rows, err := conn.Query(
			context.Background(),
			`SELECT id, number, status
			 FROM invoices
			 ORDER BY id`,
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

		invoices := []models.Invoice{}

		// Lê todas as notas fiscais
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

		// Fecha a consulta antes de fazer outra consulta
		rows.Close()

		// Consulta todos os itens das notas fiscais
		itemRows, err := conn.Query(
			context.Background(),
			`SELECT id, invoice_id, product_id, quantity
			 FROM invoice_items
			 ORDER BY invoice_id, id`,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "list_invoice_items").
				Msg("erro ao buscar itens das notas fiscais")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar itens da nota fiscal",
			})
			return
		}

		// Lê os itens
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
					Msg("erro ao ler item da nota fiscal")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao ler item da nota fiscal",
				})
				return
			}

			// Procura a nota correspondente ao item
			for i := range invoices {
				if invoices[i].ID == item.InvoiceID {
					invoices[i].Items = append(
						invoices[i].Items,
						item,
					)

					break
				}
			}
		}

		itemRows.Close()

		log.Info().
			Str("service", "invoicing").
			Str("operation", "list_invoices").
			Int("invoices_count", len(invoices)).
			Msg("notas fiscais listadas")

		// Retorna as notas fiscais com seus respectivos itens
		json.NewEncoder(c.Writer).Encode(invoices)
	}
}

// Verifica se um produto existe no Stock Service
func productExists(productID int) error {
	url := fmt.Sprintf(
		"http://localhost:8081/products/%d",
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
