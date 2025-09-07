package query

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// MockRepository é um mock do repositório de contas
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(acc *account.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockRepository) FindByID(id string) (*account.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockRepository) FindByEmail(email string) (*account.Account, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockRepository) FindAll() ([]*account.Account, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockRepository) Update(acc *account.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAccountQueryHandler_GetByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	accountID := "account-123"
	expectedAccount := &account.Account{
		ID:      accountID,
		Name:    "João Silva",
		Email:   "joao@example.com",
		Balance: 100.0,
		Status:  account.StatusActive,
	}

	// Mock: buscar conta por ID
	mockRepo.On("FindByID", accountID).Return(expectedAccount, nil)

	// Act
	result, err := handler.GetByID(accountID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, accountID, result.ID)
	assert.Equal(t, "João Silva", result.Name)
	assert.Equal(t, "joao@example.com", result.Email)
	assert.Equal(t, 100.0, result.Balance)
	assert.Equal(t, "active", result.Status)

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetByID_AccountNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	accountID := "non-existent-account"

	// Mock: conta não encontrada
	mockRepo.On("FindByID", accountID).Return(nil, errors.New("not found"))

	// Act
	result, err := handler.GetByID(accountID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetByID_AccountNotFound_NilReturn(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	accountID := "non-existent-account"

	// Mock: conta não encontrada (retorna nil)
	mockRepo.On("FindByID", accountID).Return(nil, nil)

	// Act
	result, err := handler.GetByID(accountID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrAccountNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetByEmail_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	email := "joao@example.com"
	expectedAccount := &account.Account{
		ID:      "account-123",
		Name:    "João Silva",
		Email:   email,
		Balance: 100.0,
		Status:  account.StatusActive,
	}

	// Mock: buscar conta por email
	mockRepo.On("FindByEmail", email).Return(expectedAccount, nil)

	// Act
	result, err := handler.GetByEmail(email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "account-123", result.ID)
	assert.Equal(t, "João Silva", result.Name)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, 100.0, result.Balance)
	assert.Equal(t, "active", result.Status)

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetByEmail_AccountNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	email := "nonexistent@example.com"

	// Mock: conta não encontrada
	mockRepo.On("FindByEmail", email).Return(nil, errors.New("not found"))

	// Act
	result, err := handler.GetByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetByEmail_AccountNotFound_NilReturn(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	email := "nonexistent@example.com"

	// Mock: conta não encontrada (retorna nil)
	mockRepo.On("FindByEmail", email).Return(nil, nil)

	// Act
	result, err := handler.GetByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrAccountNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetAll_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	expectedAccounts := []*account.Account{
		{
			ID:      "account-1",
			Name:    "João Silva",
			Email:   "joao@example.com",
			Balance: 100.0,
			Status:  account.StatusActive,
		},
		{
			ID:      "account-2",
			Name:    "Maria Santos",
			Email:   "maria@example.com",
			Balance: 200.0,
			Status:  account.StatusActive,
		},
	}

	// Mock: buscar todas as contas
	mockRepo.On("FindAll").Return(expectedAccounts, nil)

	// Act
	result, err := handler.GetAll()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verificar primeira conta
	assert.Equal(t, "account-1", result[0].ID)
	assert.Equal(t, "João Silva", result[0].Name)
	assert.Equal(t, "joao@example.com", result[0].Email)
	assert.Equal(t, 100.0, result[0].Balance)
	assert.Equal(t, "active", result[0].Status)

	// Verificar segunda conta
	assert.Equal(t, "account-2", result[1].ID)
	assert.Equal(t, "Maria Santos", result[1].Name)
	assert.Equal(t, "maria@example.com", result[1].Email)
	assert.Equal(t, 200.0, result[1].Balance)
	assert.Equal(t, "active", result[1].Status)

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetAll_EmptyList(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	// Mock: lista vazia (slice vazia, não nil)
	emptySlice := make([]*account.Account, 0)
	mockRepo.On("FindAll").Return(emptySlice, nil)

	// Act
	result, err := handler.GetAll()

	// Assert
	assert.NoError(t, err)
	// result pode ser nil ou slice vazia, ambos são válidos para lista vazia
	if result != nil {
		assert.Len(t, result, 0)
	}

	mockRepo.AssertExpectations(t)
}

func TestAccountQueryHandler_GetAll_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	handler := NewAccountQueryHandler(mockRepo)

	// Mock: erro no repositório
	mockRepo.On("FindAll").Return(nil, errors.New("database error"))

	// Act
	result, err := handler.GetAll()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")

	mockRepo.AssertExpectations(t)
}

func TestMapToDTO_AllStatuses(t *testing.T) {
	// Teste para verificar se o mapeamento funciona para todos os status
	testCases := []struct {
		name           string
		accountStatus  account.AccountStatus
		expectedStatus string
	}{
		{
			name:           "Active status",
			accountStatus:  account.StatusActive,
			expectedStatus: "active",
		},
		{
			name:           "Inactive status",
			accountStatus:  account.StatusInactive,
			expectedStatus: "inactive",
		},
		{
			name:           "Blocked status",
			accountStatus:  account.StatusBlocked,
			expectedStatus: "blocked",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			acc := &account.Account{
				ID:      "account-123",
				Name:    "Test User",
				Email:   "test@example.com",
				Balance: 100.0,
				Status:  tc.accountStatus,
			}

			// Act
			result := mapToDTO(acc)

			// Assert
			assert.Equal(t, "account-123", result.ID)
			assert.Equal(t, "Test User", result.Name)
			assert.Equal(t, "test@example.com", result.Email)
			assert.Equal(t, 100.0, result.Balance)
			assert.Equal(t, tc.expectedStatus, result.Status)
		})
	}
}
