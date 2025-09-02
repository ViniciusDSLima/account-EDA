package command

import "errors"

// Erros comuns para comandos
var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidAmount      = errors.New("invalid amount")
)
