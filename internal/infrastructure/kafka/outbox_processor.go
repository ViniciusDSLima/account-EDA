package kafka

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/persistence"
)

// OutboxProcessor é responsável por processar eventos do outbox
type OutboxProcessor struct {
	outboxRepo     *persistence.OutboxRepository
	publisher      *EventPublisher
	batchSize      int
	processingTime time.Duration
	maxRetries     int
	stopCh         chan struct{}
}

// NewOutboxProcessor cria um novo processador de outbox
func NewOutboxProcessor(
	outboxRepo *persistence.OutboxRepository,
	publisher *EventPublisher,
	batchSize int,
	processingTime time.Duration,
	maxRetries int,
) *OutboxProcessor {
	if batchSize <= 0 {
		batchSize = 50
	}
	if processingTime <= 0 {
		processingTime = 5 * time.Second
	}
	if maxRetries <= 0 {
		maxRetries = 5
	}

	return &OutboxProcessor{
		outboxRepo:     outboxRepo,
		publisher:      publisher,
		batchSize:      batchSize,
		processingTime: processingTime,
		maxRetries:     maxRetries,
		stopCh:         make(chan struct{}),
	}
}

// Start inicia o processador em uma goroutine
func (p *OutboxProcessor) Start() {
	go p.process()
}

// Stop interrompe o processador
func (p *OutboxProcessor) Stop() {
	close(p.stopCh)
}

// process processa eventos pendentes do outbox
func (p *OutboxProcessor) process() {
	ticker := time.NewTicker(p.processingTime)
	defer ticker.Stop()

	// Flag para controlar logs de erros repetidos
	var lastErrorMessage string
	var errorRepeatCount int

	for {
		select {
		case <-ticker.C:
			err := p.processNextBatch()
			if err != nil {
				// Implementação para evitar spam de logs com o mesmo erro
				currentError := err.Error()
				if currentError == lastErrorMessage {
					errorRepeatCount++

					// Só loga a cada 10 ocorrências do mesmo erro
					if errorRepeatCount >= 10 {
						log.Printf("Erro ao processar lote do outbox (repetido %d vezes): %v",
							errorRepeatCount, err)
						errorRepeatCount = 0
					}
				} else {
					// Novo tipo de erro, loga imediatamente
					log.Printf("Erro ao processar lote do outbox: %v", err)
					lastErrorMessage = currentError
					errorRepeatCount = 0
				}
			} else {
				// Reset do contador de erros quando um processamento bem-sucedido ocorre
				lastErrorMessage = ""
				errorRepeatCount = 0
			}
		case <-p.stopCh:
			log.Println("Processador de outbox interrompido")
			return
		}
	}
}

// processNextBatch processa o próximo lote de eventos pendentes
func (p *OutboxProcessor) processNextBatch() error {
	events, err := p.outboxRepo.GetPendingEvents(p.batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		now := time.Now().Unix()
		if now%60 == 0 {
			log.Printf("Nenhum evento pendente no outbox")
		}

		return nil
	}

	log.Printf("Processando %d eventos do outbox", len(events))

	for _, event := range events {
		// Pular eventos que excederam o número máximo de tentativas
		if event.RetryCount > p.maxRetries {
			// Enviar para DLQ
			p.handleFailedEvent(event)
			continue
		}

		// Tentar publicar o evento
		// A função publishEvent já lida com o logging de erros,
		// então não duplicamos os logs aqui
		_ = p.publishEvent(event)
	}

	return nil
}

// publishEvent publica um evento do outbox no Kafka
func (p *OutboxProcessor) publishEvent(event persistence.OutboxEvent) error {
	// Dependendo do tipo de evento, criar o objeto de domínio apropriado
	var domainEvent account.Event

	switch event.EventType {
	case "AccountCreated":
		var e account.AccountCreatedEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		domainEvent = e

	case "AccountDeposited":
		var e account.AccountDepositedEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		domainEvent = e

	case "AccountWithdrawn":
		var e account.AccountWithdrawnEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		domainEvent = e

	// Adicionar outros tipos de eventos conforme necessário

	default:
		log.Printf("Tipo de evento desconhecido: %s", event.EventType)
		return nil
	}

	// Publicar o evento
	err := p.publisher.Publish(domainEvent)

	if err != nil {
		// Verifica se o erro é de infraestrutura (broker indisponível ou tópico inexistente)
		if isKafkaInfrastructureError(err) {
			// Para erros de infraestrutura, apenas registra o erro mas não marca como falha
			// Isso evita que eventos sejam marcados como falha quando o problema é temporário
			log.Printf("Erro de infraestrutura Kafka ao processar evento %s (será tentado novamente): %v",
				event.ID, err)
			return err
		}

		// Para outros tipos de erro (formato inválido, etc), marca como falha
		if markErr := p.outboxRepo.MarkAsFailed(event.ID, err); markErr != nil {
			log.Printf("Erro ao marcar evento como falho: %v", markErr)
		}
		return err
	}

	// Marcar evento como publicado
	if err := p.outboxRepo.MarkAsPublished(event.ID); err != nil {
		log.Printf("Erro ao marcar evento como publicado: %v", err)
		return err
	}

	return nil
}

// isKafkaInfrastructureError verifica se o erro é relacionado à infraestrutura do Kafka
// e não a problemas com o evento em si
func isKafkaInfrastructureError(err error) bool {
	errorMsg := err.Error()

	// Verifica por erros comuns de infraestrutura
	infrastructureErrors := []string{
		"Unknown Topic Or Partition",
		"Leader Not Available",
		"Network Error",
		"Broker Not Available",
		"Connection Refused",
		"Connection Reset",
		"Connection Closed",
		"dial tcp",
		"i/o timeout",
		"timeout",
		"EOF",
		"broken pipe",
		"no such host",
	}

	for _, errText := range infrastructureErrors {
		if strings.Contains(errorMsg, errText) {
			return true
		}
	}

	return false
}

// handleFailedEvent lida com eventos que falharam muitas vezes
func (p *OutboxProcessor) handleFailedEvent(event persistence.OutboxEvent) {
	// Tentar enviar para a DLQ
	var domainEvent account.Event
	switch event.EventType {
	case "AccountCreated":
		var e account.AccountCreatedEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			log.Printf("Erro ao deserializar evento %s: %v", event.ID, err)
			return
		}
		domainEvent = e

	// Adicionar outros tipos conforme necessário

	default:
		log.Printf("Tipo de evento desconhecido para DLQ: %s", event.EventType)
		return
	}

	reason := "Excedeu número máximo de tentativas"
	if err := p.publisher.PublishToDLQ(domainEvent, reason); err != nil {
		log.Printf("ERRO CRÍTICO: Falha ao publicar evento %s para DLQ: %v", event.ID, err)
		return
	}

	// Mesmo que o evento tenha ido para a DLQ, ainda o marcamos como processado no outbox
	if err := p.outboxRepo.MarkAsPublished(event.ID); err != nil {
		log.Printf("Erro ao marcar evento enviado para DLQ como processado: %v", err)
	}
}
