package account

import "time"

// Event é a interface base para todos os eventos de domínio
type Event interface {
	EventName() string
	AggregateID() string
	OccurredAt() time.Time
}

// BaseEvent contém os campos comuns a todos os eventos
type BaseEvent struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	AggrID    string    `json:"aggregate_id"`
}

// EventName retorna o nome do evento
func (e BaseEvent) EventName() string {
	return e.EventType
}

// AggregateID retorna o ID do agregado
func (e BaseEvent) AggregateID() string {
	return e.AggrID
}

// OccurredAt retorna quando o evento ocorreu
func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AccountCreatedEvent é emitido quando uma conta é criada
type AccountCreatedEvent struct {
	BaseEvent
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AccountDepositedEvent é emitido quando um depósito é feito
type AccountDepositedEvent struct {
	BaseEvent
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
}

// AccountWithdrawnEvent é emitido quando um saque é feito
type AccountWithdrawnEvent struct {
	BaseEvent
	Amount         float64 `json:"amount"`
	CurrentBalance float64 `json:"current_balance"`
}

// AccountBlockedEvent é emitido quando uma conta é bloqueada
type AccountBlockedEvent struct {
	BaseEvent
	Reason string `json:"reason"`
}

// AccountActivatedEvent é emitido quando uma conta é ativada
type AccountActivatedEvent struct {
	BaseEvent
}
