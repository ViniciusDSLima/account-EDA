package command

import (
	"time"

	"github.com/google/uuid"

	"github.com/viniciuslima/account-EDA/internal/application/event"
	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// DepositCommand representa o comando para depositar em uma conta
type DepositCommand struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

// DepositHandler manipula o comando de depósito
type DepositHandler struct {
	repository account.Repository
	publisher  event.Publisher
}

// NewDepositHandler cria um novo manipulador de depósito
func NewDepositHandler(repository account.Repository, publisher event.Publisher) *DepositHandler {
	return &DepositHandler{
		repository: repository,
		publisher:  publisher,
	}
}

// Handle executa o comando de depósito
func (h *DepositHandler) Handle(cmd DepositCommand) error {
	// Validar valor do depósito
	if cmd.Amount <= 0 {
		return ErrInvalidAmount
	}

	// Buscar a conta
	acc, err := h.repository.FindByID(cmd.AccountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return ErrAccountNotFound
	}

	// Realizar o depósito
	if err := acc.Deposit(cmd.Amount); err != nil {
		return err
	}

	// Atualizar a conta
	if err := h.repository.Update(acc); err != nil {
		return err
	}

	// Publicar evento de depósito
	event := account.AccountDepositedEvent{
		BaseEvent: account.BaseEvent{
			ID:        uuid.New().String(),
			AccountID: acc.ID,
			EventType: "AccountDeposited",
			Timestamp: time.Now(),
			AggrID:    acc.ID,
		},
		Amount:         cmd.Amount,
		CurrentBalance: acc.Balance,
	}

	if err := h.publisher.Publish(event); err != nil {
		// Log o erro, mas não falha a operação
	}

	return nil
}
