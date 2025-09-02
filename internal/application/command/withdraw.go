package command

import (
	"time"

	"github.com/google/uuid"

	"github.com/viniciuslima/account-EDA/internal/application/event"
	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// WithdrawCommand representa o comando para sacar de uma conta
type WithdrawCommand struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

// WithdrawHandler manipula o comando de saque
type WithdrawHandler struct {
	repository account.Repository
	publisher  event.Publisher
}

// NewWithdrawHandler cria um novo manipulador de saque
func NewWithdrawHandler(repository account.Repository, publisher event.Publisher) *WithdrawHandler {
	return &WithdrawHandler{
		repository: repository,
		publisher:  publisher,
	}
}

// Handle executa o comando de saque
func (h *WithdrawHandler) Handle(cmd WithdrawCommand) error {
	// Validar valor do saque
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

	// Verificar se há saldo suficiente
	if acc.Balance < cmd.Amount {
		return ErrInsufficientFunds
	}

	// Realizar o saque
	if err := acc.Withdraw(cmd.Amount); err != nil {
		return err
	}

	// Atualizar a conta
	if err := h.repository.Update(acc); err != nil {
		return err
	}

	// Publicar evento de saque
	event := account.AccountWithdrawnEvent{
		BaseEvent: account.BaseEvent{
			ID:        uuid.New().String(),
			AccountID: acc.ID,
			EventType: "AccountWithdrawn",
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
