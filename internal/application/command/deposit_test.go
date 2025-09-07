package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

func TestDepositHandler_Handle_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    100.0,
	}

	existingAccount := &account.Account{
		ID:      "account-123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		Balance: 50.0,
		Status:  account.StatusActive,
	}

	// Mock: buscar conta
	mockRepo.On("FindByID", cmd.AccountID).Return(existingAccount, nil)

	// Mock: atualizar conta
	mockRepo.On("Update", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: publicar evento
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountDepositedEvent")).Return(nil)

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 150.0, existingAccount.Balance) // 50 + 100

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestDepositHandler_Handle_InvalidAmount(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    0, // Valor inválido
	}

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)
}

func TestDepositHandler_Handle_NegativeAmount(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    -50.0, // Valor negativo
	}

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)
}

func TestDepositHandler_Handle_AccountNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "non-existent-account",
		Amount:    100.0,
	}

	// Mock: conta não encontrada
	mockRepo.On("FindByID", cmd.AccountID).Return(nil, errors.New("not found"))

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	mockRepo.AssertExpectations(t)
}

func TestDepositHandler_Handle_AccountNotFound_NilReturn(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "non-existent-account",
		Amount:    100.0,
	}

	// Mock: conta não encontrada (retorna nil)
	mockRepo.On("FindByID", cmd.AccountID).Return(nil, nil)

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrAccountNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestDepositHandler_Handle_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    100.0,
	}

	// Mock: erro no repositório
	mockRepo.On("FindByID", cmd.AccountID).Return(nil, errors.New("database error"))

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockRepo.AssertExpectations(t)
}

func TestDepositHandler_Handle_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    100.0,
	}

	existingAccount := &account.Account{
		ID:      "account-123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		Balance: 50.0,
		Status:  account.StatusActive,
	}

	// Mock: buscar conta
	mockRepo.On("FindByID", cmd.AccountID).Return(existingAccount, nil)

	// Mock: erro ao atualizar conta
	mockRepo.On("Update", mock.AnythingOfType("*account.Account")).Return(errors.New("update error"))

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")

	mockRepo.AssertExpectations(t)
}

func TestDepositHandler_Handle_PublisherError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    100.0,
	}

	existingAccount := &account.Account{
		ID:      "account-123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		Balance: 50.0,
		Status:  account.StatusActive,
	}

	// Mock: buscar conta
	mockRepo.On("FindByID", cmd.AccountID).Return(existingAccount, nil)

	// Mock: atualizar conta
	mockRepo.On("Update", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: erro ao publicar evento
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountDepositedEvent")).Return(errors.New("publish error"))

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err)                          // A operação principal não deve falhar por erro de publicação
	assert.Equal(t, 150.0, existingAccount.Balance) // 50 + 100

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestDepositHandler_Handle_InactiveAccount(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)

	handler := NewDepositHandler(mockRepo, mockPublisher)

	cmd := DepositCommand{
		AccountID: "account-123",
		Amount:    100.0,
	}

	existingAccount := &account.Account{
		ID:      "account-123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		Balance: 50.0,
		Status:  account.StatusInactive, // Conta inativa
	}

	// Mock: buscar conta
	mockRepo.On("FindByID", cmd.AccountID).Return(existingAccount, nil)

	// Act
	err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account is not active")

	mockRepo.AssertExpectations(t)
}
