package main

import (
	"context"       //banco de dados
	"encoding/json" // json para requisições HTTP
	"fmt"           // biblioteca padrão para formatação de strings
	"net/http"      // biblioteca padrão para criar servidor HTTP

	"github.com/jackc/pgx/v5" // biblioteca para conectar ao PostgreSQL
)

func createNFSe(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var invoice Invoice

		err := json.NewDecoder(r.Body).Decode(&invoice)

		if err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
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
			http.Error(w, "Erro ao ao cadastrar nota fiscal", http.StatusInternalServerError)
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
				http.Error(w, "Erro ao adicionar item à nota fiscal", http.StatusInternalServerError)
				return
			}

			item.InvoiceID = invoice.ID
		}

		fmt.Printf("Nota fiscal cadastrada...: %+v\n", invoice)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(invoice)
	}
}

// Retorna a lista de notas fiscais em formato JSON
func invoicesHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Consulta todas as notas fiscais no banco de dados
		rows, err := conn.Query(
			context.Background(),
			`SELECT id, number, status
			 FROM invoices
			 ORDER BY id`,
		)

		if err != nil {
			http.Error(w, "Erro ao buscar notas fiscais", http.StatusInternalServerError)
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
				http.Error(w, "Erro ao ler nota fiscal", http.StatusInternalServerError)
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
			fmt.Println("Erro ao buscar itens da nota fiscal:", err)
			http.Error(w, "Erro ao buscar itens da nota fiscal", http.StatusInternalServerError)
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
				http.Error(w, "Erro ao ler item da nota fiscal", http.StatusInternalServerError)
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
		json.NewEncoder(w).Encode(invoices)
	}
}

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

	http.HandleFunc("/invoices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createNFSe(conn)(w, r)
		} else if r.Method == http.MethodGet {
			invoicesHandler(conn)(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Invoicing Service iniciado em http://localhost:8082")

	err = http.ListenAndServe(":8082", nil)

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}
