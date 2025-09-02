package event

import "github.com/viniciuslima/account-EDA/internal/domain/account"

// Publisher define a interface para publicação de eventos
type Publisher interface {
	// Publish publica um evento
	Publish(event account.Event) error

	// PublishToDLQ publica um evento na fila de mensagens mortas
	PublishToDLQ(event account.Event, errMsg string) error
}

// Handler define a interface para manipuladores de eventos
type Handler interface {
	// Handle processa um evento
	Handle(event account.Event) error
}
