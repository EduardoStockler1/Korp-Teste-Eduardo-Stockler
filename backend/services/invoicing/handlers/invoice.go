package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/EduardoStockler1/Korp-Teste-Eduardo-Stockler/backend/services/invoicing/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Dados do produto retornados pelo Stock Service
type StockProduct struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
}

// Cria nota fiscal e seus itens no banco de dados
func CreateNFSeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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

		// Toda NFSe nova começa como OPEN.
		// O status não é definido pelo cliente.
		invoice.Status = "OPEN"

		// Consulta o Stock Service para validar os produtos
		// e obter a descrição de cada produto.
		for i := range invoice.Items {
			item := &invoice.Items[i]

			product, err := getProduct(item.ProductID)
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

			// Guarda a descrição do produto na NFSe.
			item.ProductDescription = product.Description
		}

		// Inicia a transação usando uma conexão do pool.
		tx, err := pool.Begin(context.Background())
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

		// Se algo der errado antes do Commit, desfaz tudo.
		defer tx.Rollback(context.Background())

		// Insere a invoice.
		err = tx.QueryRow(
			context.Background(),
			`INSERT INTO invoices (number, status)
			 VALUES (NEXTVAL('invoice_number_seq'), $1)
			 RETURNING id, number`,
			invoice.Status,
		).Scan(
			&invoice.ID,
			&invoice.Number,
		)

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

		// Insere os itens da invoice.
		for i := range invoice.Items {
			item := &invoice.Items[i]

			err = tx.QueryRow(
				context.Background(),
				`INSERT INTO invoice_items (
					invoice_id,
					product_id,
					product_description,
					quantity
				)
				VALUES ($1, $2, $3, $4)
				RETURNING id`,
				invoice.ID,
				item.ProductID,
				item.ProductDescription,
				item.Quantity,
			).Scan(&item.ID)

			if err != nil {
				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "insert_invoice_item").
					Int("invoice_id", invoice.ID).
					Int("product_id", item.ProductID).
					Str("product_description", item.ProductDescription).
					Int("quantity", item.Quantity).
					Msg("erro ao inserir item da invoice")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao adicionar item à nota fiscal",
				})
				return
			}

			item.InvoiceID = invoice.ID
		}

		// Confirma a transação.
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
func InvoicesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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

		// Total de invoices
		var total int

		err = pool.QueryRow(
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

		rows, err := pool.Query(
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

		for rows.Next() {
			var invoice models.Invoice

			err := rows.Scan(
				&invoice.ID,
				&invoice.Number,
				&invoice.Status,
			)

			if err != nil {
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

			// Inicializa os itens como uma lista vazia.
			invoice.Items = []models.InvoiceItem{}

			invoices = append(invoices, invoice)
		}

		if err := rows.Err(); err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "list_invoices").
				Msg("erro ao percorrer notas fiscais")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar notas fiscais",
			})
			return
		}

		// Consulta os itens das notas fiscais.
		for i := range invoices {

			itemRows, err := pool.Query(
				context.Background(),
				`SELECT
					id,
					invoice_id,
					product_id,
					product_description,
					quantity
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
					&item.ProductDescription,
					&item.Quantity,
				)

				if err != nil {
					itemRows.Close()

					log.Error().
						Err(err).
						Str("service", "invoicing").
						Str("operation", "scan_invoice_item").
						Int("invoice_id", invoices[i].ID).
						Msg("erro ao ler item da nota fiscal")

					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Erro ao ler item da nota fiscal",
					})
					return
				}

				invoices[i].Items = append(
					invoices[i].Items,
					item,
				)
			}

			if err := itemRows.Err(); err != nil {
				itemRows.Close()

				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "list_invoice_items").
					Int("invoice_id", invoices[i].ID).
					Msg("erro ao percorrer itens da nota fiscal")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao buscar itens da nota fiscal",
				})
				return
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

// Consulta um produto no Stock Service.
// Além de validar sua existência, retorna os dados do produto.
func getProduct(productID int) (*StockProduct, error) {
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
			Str("operation", "get_product").
			Int("product_id", productID).
			Msg("erro ao consultar Stock Service")

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		err := fmt.Errorf(
			"produto %d não encontrado",
			productID,
		)

		log.Warn().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "get_product").
			Int("product_id", productID).
			Int("status_code", resp.StatusCode).
			Msg("produto não encontrado no Stock Service")

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf(
			"erro ao consultar Stock Service",
		)

		log.Error().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "get_product").
			Int("product_id", productID).
			Int("status_code", resp.StatusCode).
			Msg("Stock Service retornou erro")

		return nil, err
	}

	var product StockProduct

	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		log.Error().
			Err(err).
			Str("service", "invoicing").
			Str("operation", "get_product").
			Int("product_id", productID).
			Msg("erro ao interpretar produto do Stock Service")

		return nil, err
	}

	return &product, nil
}

// Imprime uma nota fiscal, baixa o estoque dos produtos e fecha a nota.
func PrintInvoiceHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Warn().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Str("invoice_id", c.Param("id")).
				Msg("ID da nota fiscal inválido")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID da nota fiscal inválido",
			})
			return
		}

		var status string

		err = pool.QueryRow(
			context.Background(),
			`SELECT status
			 FROM invoices
			 WHERE id = $1`,
			invoiceID,
		).Scan(&status)

		if err != nil {
			if err == pgx.ErrNoRows {
				log.Warn().
					Str("service", "invoicing").
					Str("operation", "print_invoice").
					Int("invoice_id", invoiceID).
					Msg("nota fiscal não encontrada")

				c.JSON(http.StatusNotFound, gin.H{
					"error": "Nota fiscal não encontrada",
				})
				return
			}

			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Int("invoice_id", invoiceID).
				Msg("erro ao buscar nota fiscal")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar nota fiscal",
			})
			return
		}

		// Se já estiver fechada, não baixa o estoque novamente.
		if status == "CLOSED" {
			log.Info().
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Int("invoice_id", invoiceID).
				Msg("nota fiscal já estava fechada")

			c.JSON(http.StatusOK, gin.H{
				"message": "Nota fiscal já foi impressa",
			})
			return
		}

		if status != "OPEN" {
			log.Warn().
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Int("invoice_id", invoiceID).
				Str("status", status).
				Msg("nota fiscal não pode ser impressa")

			c.JSON(http.StatusConflict, gin.H{
				"error": "Nota fiscal não pode ser impressa",
			})
			return
		}

		rows, err := pool.Query(
			context.Background(),
			`SELECT product_id, quantity
			 FROM invoice_items
			 WHERE invoice_id = $1
			 ORDER BY id`,
			invoiceID,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Int("invoice_id", invoiceID).
				Msg("erro ao buscar itens da nota fiscal")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar itens da nota fiscal",
			})
			return
		}

		type InvoiceStockItem struct {
			ProductID int
			Quantity  int
		}

		items := []InvoiceStockItem{}

		for rows.Next() {
			var item InvoiceStockItem

			if err := rows.Scan(
				&item.ProductID,
				&item.Quantity,
			); err != nil {
				rows.Close()

				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "print_invoice").
					Int("invoice_id", invoiceID).
					Msg("erro ao ler item da nota fiscal")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Erro ao ler item da nota fiscal",
				})
				return
			}

			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			rows.Close()

			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "print_invoice").
				Int("invoice_id", invoiceID).
				Msg("erro ao percorrer itens da nota fiscal")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao buscar itens da nota fiscal",
			})
			return
		}

		rows.Close()

		// Diminui o estoque de cada produto no Stock Service.
		for _, item := range items {
			if err := decreaseProductStock(item.ProductID, item.Quantity); err != nil {
				log.Error().
					Err(err).
					Str("service", "invoicing").
					Str("operation", "decrease_stock").
					Int("invoice_id", invoiceID).
					Int("product_id", item.ProductID).
					Int("quantity", item.Quantity).
					Msg("erro ao diminuir estoque")

				c.JSON(http.StatusConflict, gin.H{
					"error": "Não foi possível diminuir o estoque",
				})
				return
			}
		}

		// Fecha a nota somente depois que todas as baixas foram realizadas.
		_, err = pool.Exec(
			context.Background(),
			`UPDATE invoices
			 SET status = 'CLOSED'
			 WHERE id = $1
			   AND status = 'OPEN'`,
			invoiceID,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("service", "invoicing").
				Str("operation", "close_invoice").
				Int("invoice_id", invoiceID).
				Msg("erro ao fechar nota fiscal")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao fechar nota fiscal",
			})
			return
		}

		log.Info().
			Str("service", "invoicing").
			Str("operation", "print_invoice").
			Int("invoice_id", invoiceID).
			Str("status", "CLOSED").
			Msg("nota fiscal impressa e estoque atualizado")

		c.JSON(http.StatusOK, gin.H{
			"message": "Nota fiscal impressa com sucesso",
			"status":  "CLOSED",
		})
	}
}

// Diminui o estoque de um produto no Stock Service.
func decreaseProductStock(productID int, quantity int) error {
	stockServiceURL := os.Getenv("STOCK_SERVICE_URL")

	if stockServiceURL == "" {
		stockServiceURL = "http://localhost:8081"
	}

	url := fmt.Sprintf(
		"%s/products/%d/decrease-stock",
		stockServiceURL,
		productID,
	)

	requestBody := struct {
		Quantity int `json:"quantity"`
	}{
		Quantity: quantity,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"Stock Service retornou status %d",
			resp.StatusCode,
		)
	}

	return nil
}
