package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

const (
	DLQSuffix = "-dlq"
)

// EventPublisher implementa event.Publisher usando Kafka
type EventPublisher struct {
	writer    *kafka.Writer
	dlqWriter *kafka.Writer
	brokers   []string
	topic     string
}

// NewEventPublisher cria um novo publicador de eventos
func NewEventPublisher(brokers []string) *EventPublisher {
	topic := "account-events"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: int(kafka.RequireAll), // -1, exige ack de todos os replicas
		MaxAttempts:  3,                     // Número de tentativas
		ReadTimeout:  5 * time.Second,       // Timeout de leitura
		WriteTimeout: 5 * time.Second,       // Timeout de escrita
		BatchBytes:   1048576,               // 1MB

	})

	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic + DLQSuffix,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: int(kafka.RequireAll),
		BatchBytes:   1048576, // 1MB

	})

	return &EventPublisher{
		writer:    writer,
		dlqWriter: dlqWriter,
		brokers:   brokers,
		topic:     topic,
	}
}

// Close fecha a conexão com o Kafka
func (p *EventPublisher) Close() error {
	if err := p.writer.Close(); err != nil {
		return err
	}
	return p.dlqWriter.Close()
}

// Publish publica um evento no Kafka
func (p *EventPublisher) Publish(event account.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateID()),
		Value: value,
		Time:  event.OccurredAt(),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventName())},
		},
	}

	// Define um timeout curto para não bloquear a aplicação quando Kafka não está disponível
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("falha ao publicar evento no Kafka: %w", err)
	}

	return nil
}

// PublishToDLQ publica um evento na fila de mensagens mortas (DLQ) com informações do erro
func (p *EventPublisher) PublishToDLQ(event account.Event, errMsg string) error {
	// Se o Kafka não estiver disponível, apenas loga o erro e continua
	// Esta implementação protege a aplicação de falhas causadas pela indisponibilidade do Kafka
	// O evento já está salvo no outbox e será processado posteriormente

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event for DLQ: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateID()),
		Value: value,
		Time:  event.OccurredAt(),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventName())},
			{Key: "error", Value: []byte(errMsg)},
			{Key: "original_topic", Value: []byte(p.topic)},
			{Key: "failure_time", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	// Define um timeout curto para não bloquear a aplicação quando Kafka não está disponível
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Tenta enviar para DLQ, mas não bloqueia a aplicação em caso de falha
	err = p.dlqWriter.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Não foi possível enviar evento para DLQ (será processado pelo outbox): %v", err)
	}

	// Sempre retorna nil para não propagar erros do Kafka para a aplicação
	// O evento já está salvo no outbox e será processado posteriormente
	return nil
}

// CreateTopics cria os tópicos necessários se não existirem
func (p *EventPublisher) CreateTopics() error {
	conn, err := kafka.Dial("tcp", p.brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topics := []string{p.topic, p.topic + DLQSuffix}

	for _, topic := range topics {
		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     3,
				ReplicationFactor: 1,
			},
		}

		err = controllerConn.CreateTopics(topicConfigs...)
		if err != nil {
			log.Printf("Aviso ao criar tópico %s: %v", topic, err)
		}
	}

	return nil
}
