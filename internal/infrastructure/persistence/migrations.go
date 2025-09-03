package persistence

import (
	"database/sql"
)

// RunMigrations executa as migrações do banco de dados
func RunMigrations(db *sql.DB) error {
	if err := createAccountsTable(db); err != nil {
		return err
	}

	if err := createOutboxTable(db); err != nil {
		return err
	}

	return nil
}

// createAccountsTable cria a tabela de contas
func createAccountsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL UNIQUE,
			balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(query)
	return err
}



// createOutboxTable cria a tabela outbox para publicação confiável de eventos
func createOutboxTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS outbox_events (
			id VARCHAR(36) PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			aggregate_id VARCHAR(36) NOT NULL,
			payload JSONB NOT NULL,
			status VARCHAR(20) NOT NULL,
			retry_count INT NOT NULL DEFAULT 0,
			error TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(query)
	return err
}
