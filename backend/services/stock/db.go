package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// conectando ao banco de dados
func connectDatabase() (*pgx.Conn, error) {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "nfs_issuer"
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	conn, err := pgx.Connect(
		context.Background(),
		databaseURL,
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
