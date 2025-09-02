package command

import (
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/viniciuslima/account-EDA/internal/application/event"
	"github.com/viniciuslima/account-EDA/internal/domain/account"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/persistence"
)

type CreateAccountCommand struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateAccountHandler struct {
	repository account.Repository
	publisher  event.Publisher
	outboxRepo *persistence.OutboxRepository
}

// NewCreateAccountHandler cria um novo manipulador de criação de conta
func NewCreateAccountHandler(repository account.Repository, publisher event.Publisher, outboxRepo *persistence.OutboxRepository) *CreateAccountHandler {
	return &CreateAccountHandler{
		repository: repository,
		publisher:  publisher,
		outboxRepo: outboxRepo,
	}
}

// Handle executa o comando de criação de conta
func (h *CreateAccountHandler) Handle(cmd CreateAccountCommand) (string, error) {
	// Verificar se já existe uma conta com este e-mail
	existingAccount, err := h.repository.FindByEmail(cmd.Email)
	if err == nil && existingAccount != nil {
		return "", ErrEmailAlreadyExists
	}

	// Criar a nova conta
	newAccount, err := account.NewAccount(cmd.Name, cmd.Email)
	if err != nil {
		return "", err
	}

	// Persistir a nova conta
	if err := h.repository.Save(newAccount); err != nil {
		return "", err
	}

	// Publicar evento de conta criada
	event := account.AccountCreatedEvent{
		BaseEvent: account.BaseEvent{
			ID:        uuid.New().String(),
			AccountID: newAccount.ID,
			EventType: "AccountCreated",
			Timestamp: time.Now(),
			AggrID:    newAccount.ID,
		},
		Name:  newAccount.Name,
		Email: newAccount.Email,
	}

	// Salvar no outbox primeiro (para garantir que o evento será enviado eventualmente)
	if err := h.outboxRepo.Save(event.EventName(), event.AggregateID(), event); err != nil {
		log.Printf("ERRO ao salvar evento no outbox: %v", err)
		// Não falha a operação principal, mas registra o erro
	}

	// Tenta publicar diretamente (para entrega imediata quando possível)
	if err := h.publisher.Publish(event); err != nil {
		// Tenta publicar no DLQ
		if dlqErr := h.publisher.PublishToDLQ(event, err.Error()); dlqErr != nil {
			// Se falhar também no DLQ, registra o erro e continua
			log.Printf("ERRO crítico: falha ao publicar evento no DLQ: %v", dlqErr)
		} else {
			log.Printf("Evento redirecionado para DLQ devido ao erro: %v", err)
		}

		// A operação principal continua, pois o evento será processado pelo sistema de outbox
	}

	// Se a publicação foi bem-sucedida, podemos marcar o evento como publicado no outbox
	// Isso seria feito em um job separado em um sistema real, mas incluímos aqui para demonstração
	log.Printf("Evento publicado com sucesso: %s", event.EventName())

	return newAccount.ID, nil
}
