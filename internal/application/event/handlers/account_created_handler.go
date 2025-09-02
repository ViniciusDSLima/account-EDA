package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// AccountCreatedHandler processa eventos de criação de conta
type AccountCreatedHandler struct {
	// Aqui você pode injetar dependências como serviços de email,
	// outros repositórios, clientes HTTP, etc.
}

// NewAccountCreatedHandler cria um novo handler para eventos de conta criada
func NewAccountCreatedHandler() *AccountCreatedHandler {
	return &AccountCreatedHandler{}
}

// EventType retorna o tipo de evento que este handler processa
func (h *AccountCreatedHandler) EventType() string {
	return "AccountCreated"
}

// Handle processa o evento de conta criada
func (h *AccountCreatedHandler) Handle(ctx context.Context, eventData []byte) error {
	var event account.AccountCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	log.Printf("Processando conta criada - ID: %s, Nome: %s, Email: %s",
		event.AccountID, event.Name, event.Email)

	// Aqui você pode adicionar lógica de negócio como:

	// 1. Enviar email de boas-vindas
	// if err := h.emailService.SendWelcomeEmail(event.Email, event.Name); err != nil {
	//     return fmt.Errorf("erro ao enviar email de boas-vindas: %w", err)
	// }

	// 2. Criar perfil em outro serviço
	// if err := h.profileService.CreateProfile(event.AccountID, event.Name); err != nil {
	//     return fmt.Errorf("erro ao criar perfil: %w", err)
	// }

	// 3. Enviar notificação para sistema de analytics
	// h.analytics.TrackUserSignup(event.AccountID, event.Email)

	// 4. Atualizar cache
	// h.cache.InvalidateUserList()

	// 5. Publicar em webhook externo
	// h.webhookClient.Notify("user.created", event)

	log.Printf("Conta %s processada com sucesso", event.AccountID)
	return nil
}
