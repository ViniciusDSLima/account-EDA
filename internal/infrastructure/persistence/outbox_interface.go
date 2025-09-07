package persistence

// OutboxRepositoryInterface define a interface para o repositório de outbox
type OutboxRepositoryInterface interface {
	// Save salva um evento no outbox
	Save(eventType, aggregateID string, payload interface{}) error

	// GetPendingEvents retorna eventos pendentes para publicação
	GetPendingEvents(limit int) ([]OutboxEvent, error)

	// MarkAsPublished marca um evento como publicado com sucesso
	MarkAsPublished(id string) error

	// MarkAsFailed marca um evento como falho, incrementando a contagem de retentativas
	MarkAsFailed(id string, err error) error
}
