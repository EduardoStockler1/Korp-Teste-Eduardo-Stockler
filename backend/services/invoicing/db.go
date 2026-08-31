package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// conectando ao banco de dados
func connectDatabase() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		context.Background(),
		"postgres://postgres:postgres@localhost:5432/nfs_issuer",
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Testando conexão com o db...
func testDatabaseConnection(conn *pgx.Conn) error {
	err := conn.Ping(context.Background())

	if err != nil {
		return err
	}

	fmt.Println("Conectado ao PostgreSQL!")

	return nil
}
