package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OutboxStatus representa o status de um evento no outbox
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
)

// OutboxEvent representa um evento armazenado no outbox para publicação confiável
type OutboxEvent struct {
	ID          string       `json:"id"`
	EventType   string       `json:"event_type"`
	AggregateID string       `json:"aggregate_id"`
	Payload     []byte       `json:"payload"`
	Status      OutboxStatus `json:"status"`
	RetryCount  int          `json:"retry_count"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// OutboxRepository é responsável pela persistência de eventos no outbox
type OutboxRepository struct {
	db *sql.DB
}

// NewOutboxRepository cria um novo repositório para o outbox
func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{
		db: db,
	}
}

// Save salva um evento no outbox
func (r *OutboxRepository) Save(eventType, aggregateID string, payload interface{}) error {
	// Serializar o payload para JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento para outbox: %w", err)
	}

	now := time.Now()
	query := `
		INSERT INTO outbox_events 
		(id, event_type, aggregate_id, payload, status, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Exec(
		query,
		uuid.New().String(),
		eventType,
		aggregateID,
		data,
		string(OutboxStatusPending),
		0,
		now,
		now,
	)

	return err
}

// GetPendingEvents retorna eventos pendentes para publicação
func (r *OutboxRepository) GetPendingEvents(limit int) ([]OutboxEvent, error) {
	query := `
		SELECT id, event_type, aggregate_id, payload, status, retry_count, error, created_at, updated_at
		FROM outbox_events
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(query, string(OutboxStatusPending), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var event OutboxEvent
		var status string
		var errorMsg sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.Payload,
			&status,
			&event.RetryCount,
			&errorMsg,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if errorMsg.Valid {
			event.Error = errorMsg.String
		}

		if err != nil {
			return nil, err
		}

		event.Status = OutboxStatus(status)
		events = append(events, event)
	}

	return events, rows.Err()
}

// MarkAsPublished marca um evento como publicado com sucesso
func (r *OutboxRepository) MarkAsPublished(id string) error {
	query := `
		UPDATE outbox_events
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, string(OutboxStatusPublished), time.Now(), id)
	return err
}

// MarkAsFailed marca um evento como falho, incrementando a contagem de retentativas
func (r *OutboxRepository) MarkAsFailed(id string, err error) error {
	query := `
		UPDATE outbox_events
		SET status = $1, retry_count = retry_count + 1, error = $2, updated_at = $3
		WHERE id = $4
	`

	_, execErr := r.db.Exec(
		query,
		string(OutboxStatusFailed),
		err.Error(),
		time.Now(),
		id,
	)

	return execErr
}
