package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Driver PostgreSQL
	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// PostgresRepository implementa account.Repository usando PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository cria um novo repositório PostgreSQL
func NewPostgresRepository(connectionString string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Verificar conexão
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

// Close fecha a conexão com o banco de dados
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// Save persiste uma conta no banco de dados
func (r *PostgresRepository) Save(account *account.Account) error {
	query := `
		INSERT INTO accounts (id, name, email, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(
		query,
		account.ID,
		account.Name,
		account.Email,
		account.Balance,
		account.Status,
		account.CreatedAt,
		account.UpdatedAt,
	)

	return err
}

// FindByID busca uma conta pelo ID
func (r *PostgresRepository) FindByID(id string) (*account.Account, error) {
	query := `
		SELECT id, name, email, balance, status, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	return r.scanAccount(row)
}

// FindByEmail busca uma conta pelo email
func (r *PostgresRepository) FindByEmail(email string) (*account.Account, error) {
	query := `
		SELECT id, name, email, balance, status, created_at, updated_at
		FROM accounts
		WHERE email = $1
	`
	row := r.db.QueryRow(query, email)

	return r.scanAccount(row)
}

// FindAll busca todas as contas
func (r *PostgresRepository) FindAll() ([]*account.Account, error) {
	query := `
		SELECT id, name, email, balance, status, created_at, updated_at
		FROM accounts
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		account, err := r.scanAccountFromRows(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// Update atualiza uma conta
func (r *PostgresRepository) Update(account *account.Account) error {
	query := `
		UPDATE accounts
		SET name = $1, email = $2, balance = $3, status = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.Exec(
		query,
		account.Name,
		account.Email,
		account.Balance,
		account.Status,
		time.Now(),
		account.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found", account.ID)
	}

	return nil
}

// Delete remove uma conta
func (r *PostgresRepository) Delete(id string) error {
	query := `DELETE FROM accounts WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found", id)
	}

	return nil
}

// scanAccount escaneia uma linha da consulta para uma entidade Account
func (r *PostgresRepository) scanAccount(row *sql.Row) (*account.Account, error) {
	var acc account.Account
	var status string

	err := row.Scan(
		&acc.ID,
		&acc.Name,
		&acc.Email,
		&acc.Balance,
		&status,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Nenhuma conta encontrada
		}
		return nil, err
	}

	acc.Status = account.AccountStatus(status)
	return &acc, nil
}

// scanAccountFromRows escaneia uma linha do resultado para uma entidade Account
func (r *PostgresRepository) scanAccountFromRows(rows *sql.Rows) (*account.Account, error) {
	var acc account.Account
	var status string

	err := rows.Scan(
		&acc.ID,
		&acc.Name,
		&acc.Email,
		&acc.Balance,
		&status,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	acc.Status = account.AccountStatus(status)
	return &acc, nil
}
