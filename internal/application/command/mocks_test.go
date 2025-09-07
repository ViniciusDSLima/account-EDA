package command

import (
	"github.com/stretchr/testify/mock"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/persistence"
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

// MockPublisher é um mock do publisher de eventos
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(event account.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockPublisher) PublishToDLQ(event account.Event, reason string) error {
	args := m.Called(event, reason)
	return args.Error(0)
}

// MockOutboxRepository é um mock do repositório de outbox
type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) Save(eventName, aggregateID string, event interface{}) error {
	args := m.Called(eventName, aggregateID, event)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetPendingEvents(limit int) ([]persistence.OutboxEvent, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]persistence.OutboxEvent), args.Error(1)
}

func (m *MockOutboxRepository) MarkAsPublished(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOutboxRepository) MarkAsFailed(id string, err error) error {
	args := m.Called(id, err)
	return args.Error(0)
}
