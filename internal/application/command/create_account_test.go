package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

func TestCreateAccountHandler_Handle_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Mock: salvar conta com sucesso
	mockRepo.On("Save", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: salvar no outbox com sucesso
	mockOutbox.On("Save", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil)

	// Mock: publicar evento com sucesso
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountCreatedEvent")).Return(nil)

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accountID)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	existingAccount := &account.Account{
		ID:    "existing-id",
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: conta já existe
	mockRepo.On("FindByEmail", cmd.Email).Return(existingAccount, nil)

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrEmailAlreadyExists, err)
	assert.Empty(t, accountID)

	mockRepo.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_InvalidAccountData(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "", // Nome vazio
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, accountID)
	assert.Contains(t, err.Error(), "name is required")

	mockRepo.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_RepositorySaveError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Mock: erro ao salvar conta
	mockRepo.On("Save", mock.AnythingOfType("*account.Account")).Return(errors.New("database error"))

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, accountID)
	assert.Contains(t, err.Error(), "database error")

	mockRepo.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_OutboxSaveError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Mock: salvar conta com sucesso
	mockRepo.On("Save", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: erro ao salvar no outbox
	mockOutbox.On("Save", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(errors.New("outbox error"))

	// Mock: publicar evento com sucesso
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountCreatedEvent")).Return(nil)

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err) // A operação principal não deve falhar
	assert.NotEmpty(t, accountID)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_PublisherError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Mock: salvar conta com sucesso
	mockRepo.On("Save", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: salvar no outbox com sucesso
	mockOutbox.On("Save", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil)

	// Mock: erro ao publicar evento
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountCreatedEvent")).Return(errors.New("publish error"))

	// Mock: publicar no DLQ com sucesso
	mockPublisher.On("PublishToDLQ", mock.AnythingOfType("account.AccountCreatedEvent"), "publish error").Return(nil)

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err) // A operação principal não deve falhar
	assert.NotEmpty(t, accountID)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

func TestCreateAccountHandler_Handle_PublisherAndDLQError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepository)

	handler := NewCreateAccountHandler(mockRepo, mockPublisher, mockOutbox)

	cmd := CreateAccountCommand{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Mock: não existe conta com este email
	mockRepo.On("FindByEmail", cmd.Email).Return(nil, errors.New("not found"))

	// Mock: salvar conta com sucesso
	mockRepo.On("Save", mock.AnythingOfType("*account.Account")).Return(nil)

	// Mock: salvar no outbox com sucesso
	mockOutbox.On("Save", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil)

	// Mock: erro ao publicar evento
	mockPublisher.On("Publish", mock.AnythingOfType("account.AccountCreatedEvent")).Return(errors.New("publish error"))

	// Mock: erro ao publicar no DLQ
	mockPublisher.On("PublishToDLQ", mock.AnythingOfType("account.AccountCreatedEvent"), "publish error").Return(errors.New("dlq error"))

	// Act
	accountID, err := handler.Handle(cmd)

	// Assert
	assert.NoError(t, err) // A operação principal não deve falhar
	assert.NotEmpty(t, accountID)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}
