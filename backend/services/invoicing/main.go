package main

import (
	"context"       //banco de dados
	"encoding/json" // json para requisições HTTP
	"fmt"           // biblioteca padrão para formatação de strings
	"net/http"

	"github.com/gin-gonic/gin" // biblioteca para criar servidor HTTP com rotas usando gin framework
	"github.com/jackc/pgx/v5"  // biblioteca para conectar ao PostgreSQL
)

func main() {
	conn, err := connectDatabase()

	if err != nil {
		fmt.Println("Erro ao conectar ao banco:", err)
		return
	}

	defer conn.Close(context.Background())

	err = testDatabaseConnection(conn)

	if err != nil {
		fmt.Println("Erro ao testar conexão:", err)
		return
	}

	r := gin.Default()
	r.POST("/invoices", createNFSe(conn))
	r.GET("/invoices", invoicesHandler(conn))

	fmt.Println("Invoicing Service iniciado em http://localhost:8082")

	err = r.Run(":8082")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}

func createNFSe(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice Invoice

		var err error

		// Lê e valida o JSON enviado na requisição
		if err := c.ShouldBindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido LINHA 51"})
			return
		}

		for _, item := range invoice.Items {

			err := productExists(item.ProductID)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao buscar produto no stock service"})
				return
			}

			return
		}

		err = conn.QueryRow(
			context.Background(),
			`INSERT INTO invoices (number, status)
			VALUES ($1, $2)
			RETURNING id`,
			invoice.Number,
			invoice.Status,
		).Scan(&invoice.ID)

		if err != nil {
			fmt.Println("ERRO AO INSERIR INVOICE:", err)

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ao cadastrar nota fiscal LINHA 79"})
			return
		}

		for i := range invoice.Items {
			item := &invoice.Items[i]

			err = conn.QueryRow(
				context.Background(),
				`INSERT INTO invoice_items (invoice_id, product_id, quantity)
		 		 VALUES ($1, $2, $3)
		 		 	RETURNING id`,
				invoice.ID,
				item.ProductID,
				item.Quantity,
			).Scan(&item.ID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar item à nota fiscal"})
				return
			}

			item.InvoiceID = invoice.ID
		}

		fmt.Printf("Nota fiscal cadastrada...: %+v\n", invoice)

		c.Header("Content-Type", "application/json")
		c.Status(http.StatusCreated)

		json.NewEncoder(c.Writer).Encode(invoice)
	}
}

// Retorna a lista de notas fiscais em formato JSON
func invoicesHandler(conn *pgx.Conn) gin.HandlerFunc {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar notas fiscais"})
			return
		}

		invoices := []Invoice{}

		// Lê todas as notas fiscais
		for rows.Next() {
			var invoice Invoice

			err := rows.Scan(
				&invoice.ID,
				&invoice.Number,
				&invoice.Status,
			)

			if err != nil {
				rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler nota fiscal LINHA 168"})
				return
			}

			// Inicializa os itens como uma lista vazia
			invoice.Items = []InvoiceItem{}

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar itens da nota fiscal"})
			return
		}

		// Lê os itens
		for itemRows.Next() {
			var item InvoiceItem

			err := itemRows.Scan(
				&item.ID,
				&item.InvoiceID,
				&item.ProductID,
				&item.Quantity,
			)

			if err != nil {
				itemRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler item da nota fiscal LINHA 190"})
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

		// Retorna as notas fiscais com seus respectivos itens
		json.NewEncoder(c.Writer).Encode(invoices)
	}
}

func productExists(productID int) error {
	url := fmt.Sprintf(
		"http://localhost:8081/products/%d",
		productID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("produto %d não encontrado", productID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao consultar Stock Service")
	}

	return nil
}
