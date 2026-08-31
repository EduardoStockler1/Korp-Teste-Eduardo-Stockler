package main

import (
	"context"       //banco de dados
	"encoding/json" //json para requisições HTTP
	"fmt"           // biblioteca padrão para formatação de strings
	"net/http"      // biblioteca padrão para criar servidor HTTP

	"github.com/jackc/pgx/v5" // biblioteca para conectar ao PostgreSQL
)

// Cria produto a patir de uma requisição HTTP POST com JSON no corpo da requisição
func createProductHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var product Product

		err := json.NewDecoder(r.Body).Decode(&product)

		if err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		err = conn.QueryRow(
			context.Background(),
			`INSERT INTO products (code, description, stock)
			 VALUES ($1, $2, $3)
			 RETURNING id`,
			product.Code,
			product.Description,
			product.Stock,
		).Scan(&product.ID)

		if err != nil {
			http.Error(w, "Erro ao ao cadastrar produto", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Produto cadastrado...: %+v\n", product)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(product)
	}
}

// Retorna a lista de produtos em formato JSON
func productsHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rows, err := conn.Query(
			context.Background(),
			`SELECT id, code, description, stock
			 FROM products
			 ORDER BY id`,
		)

		if err != nil {
			http.Error(w, "Erro ao buscar produtos", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		products := []Product{}

		for rows.Next() {
			var product Product

			err := rows.Scan(
				&product.ID,
				&product.Code,
				&product.Description,
				&product.Stock,
			)

			if err != nil {
				http.Error(w, "Erro ao ler produto", http.StatusInternalServerError)
				return
			}

			products = append(products, product)
		}

		json.NewEncoder(w).Encode(products)
	}
}

// Inicia o servidor HTTP na porta 8081
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

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			productsHandler(conn)(w, r)
		} else if r.Method == http.MethodPost {
			createProductHandler(conn)(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Stock Service iniciado em http://localhost:8081")

	err = http.ListenAndServe(":8081", nil)

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}
