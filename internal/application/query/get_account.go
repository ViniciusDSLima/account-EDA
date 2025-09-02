package query

import "github.com/viniciuslima/account-EDA/internal/domain/account"

// AccountQuery representa o serviço de consulta para contas
type AccountQuery interface {
	// GetByID busca uma conta pelo ID
	GetByID(id string) (*AccountDTO, error)

	// GetByEmail busca uma conta pelo email
	GetByEmail(email string) (*AccountDTO, error)

	// GetAll busca todas as contas
	GetAll() ([]*AccountDTO, error)
}

// AccountDTO é o objeto de transferência de dados para a entidade Account
type AccountDTO struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

// AccountQueryHandler implementa AccountQuery
type AccountQueryHandler struct {
	repository account.Repository
}

// NewAccountQueryHandler cria um novo manipulador de consultas de conta
func NewAccountQueryHandler(repository account.Repository) *AccountQueryHandler {
	return &AccountQueryHandler{
		repository: repository,
	}
}

// GetByID busca uma conta pelo ID
func (h *AccountQueryHandler) GetByID(id string) (*AccountDTO, error) {
	acc, err := h.repository.FindByID(id)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, ErrAccountNotFound
	}

	return mapToDTO(acc), nil
}

// GetByEmail busca uma conta pelo email
func (h *AccountQueryHandler) GetByEmail(email string) (*AccountDTO, error) {
	acc, err := h.repository.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, ErrAccountNotFound
	}

	return mapToDTO(acc), nil
}

// GetAll busca todas as contas
func (h *AccountQueryHandler) GetAll() ([]*AccountDTO, error) {
	accounts, err := h.repository.FindAll()
	if err != nil {
		return nil, err
	}

	var dtos []*AccountDTO
	for _, acc := range accounts {
		dtos = append(dtos, mapToDTO(acc))
	}

	return dtos, nil
}

// mapToDTO converte uma entidade Account para AccountDTO
func mapToDTO(acc *account.Account) *AccountDTO {
	return &AccountDTO{
		ID:      acc.ID,
		Name:    acc.Name,
		Email:   acc.Email,
		Balance: acc.Balance,
		Status:  string(acc.Status),
	}
}
