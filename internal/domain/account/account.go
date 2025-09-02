package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Account representa a entidade de domínio Conta
type Account struct {
	ID        string
	Name      string
	Email     string
	Balance   float64
	Status    AccountStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountStatus é o tipo para o status da conta
type AccountStatus string

const (
	StatusActive   AccountStatus = "active"
	StatusInactive AccountStatus = "inactive"
	StatusBlocked  AccountStatus = "blocked"
)

func NewAccount(name, email string) (*Account, error) {
	if err := validateAccount(name, email); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Account{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		Balance:   0,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func validateAccount(name, email string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if email == "" {
		return errors.New("email is required")
	}
	return nil
}

func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}
	if a.Status != StatusActive {
		return errors.New("account is not active")
	}

	a.Balance += amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("withdraw amount must be positive")
	}
	if a.Status != StatusActive {
		return errors.New("account is not active")
	}
	if a.Balance < amount {
		return errors.New("insufficient funds")
	}

	a.Balance -= amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Block() error {
	if a.Status == StatusBlocked {
		return errors.New("account is already blocked")
	}
	a.Status = StatusBlocked
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Activate() error {
	if a.Status == StatusActive {
		return errors.New("account is already active")
	}
	a.Status = StatusActive
	a.UpdatedAt = time.Now()
	return nil
}
