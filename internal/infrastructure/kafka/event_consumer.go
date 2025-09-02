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

// EventHandler define a interface para manipuladores de eventos específicos
type EventHandler interface {
	Handle(ctx context.Context, event []byte) error
	EventType() string
}

// EventConsumer consome eventos do Kafka e os processa
type EventConsumer struct {
	reader   *kafka.Reader
	handlers map[string]EventHandler
	stopCh   chan struct{}
}

// NewEventConsumer cria um novo consumidor de eventos
func NewEventConsumer(brokers []string, groupID string, topic string) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})

	return &EventConsumer{
		reader:   reader,
		handlers: make(map[string]EventHandler),
		stopCh:   make(chan struct{}),
	}
}

// RegisterHandler registra um manipulador para um tipo específico de evento
func (c *EventConsumer) RegisterHandler(handler EventHandler) {
	c.handlers[handler.EventType()] = handler
	log.Printf("Registrado handler para evento: %s", handler.EventType())
}

// Start inicia o consumo de mensagens
func (c *EventConsumer) Start(ctx context.Context) error {
	log.Println("Iniciando consumidor de eventos...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopCh:
			return nil
		default:
			// Lê a próxima mensagem
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Erro ao buscar mensagem: %v", err)
				time.Sleep(time.Second)
				continue
			}

			// Processa a mensagem
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Erro ao processar mensagem: %v", err)
				// Em caso de erro, não commita o offset
				continue
			}

			// Commita o offset apenas após processar com sucesso
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Erro ao commitar mensagem: %v", err)
			}
		}
	}
}

// Stop para o consumidor
func (c *EventConsumer) Stop() error {
	close(c.stopCh)
	return c.reader.Close()
}

// processMessage processa uma mensagem individual
func (c *EventConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	// Extrai o tipo de evento dos headers
	var eventType string
	for _, header := range msg.Headers {
		if header.Key == "event_type" {
			eventType = string(header.Value)
			break
		}
	}

	if eventType == "" {
		// Tenta extrair do JSON se não estiver no header
		var baseEvent account.BaseEvent
		if err := json.Unmarshal(msg.Value, &baseEvent); err != nil {
			return fmt.Errorf("erro ao extrair tipo de evento: %w", err)
		}
		eventType = baseEvent.EventType
	}

	log.Printf("Processando evento: %s (offset: %d, partition: %d)",
		eventType, msg.Offset, msg.Partition)

	// Busca o handler apropriado
	handler, exists := c.handlers[eventType]
	if !exists {
		log.Printf("Nenhum handler registrado para evento: %s", eventType)
		// Retorna nil para não bloquear o consumo
		return nil
	}

	// Processa o evento com o handler específico
	if err := handler.Handle(ctx, msg.Value); err != nil {
		return fmt.Errorf("erro no handler do evento %s: %w", eventType, err)
	}

	log.Printf("Evento %s processado com sucesso", eventType)
	return nil
}
