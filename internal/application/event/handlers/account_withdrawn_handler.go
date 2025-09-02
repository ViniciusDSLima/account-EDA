package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// AccountWithdrawnHandler processa eventos de saque em conta
type AccountWithdrawnHandler struct {
	// Dependências como serviços de notificação, limites, etc.
}

// NewAccountWithdrawnHandler cria um novo handler para eventos de saque
func NewAccountWithdrawnHandler() *AccountWithdrawnHandler {
	return &AccountWithdrawnHandler{}
}

// EventType retorna o tipo de evento que este handler processa
func (h *AccountWithdrawnHandler) EventType() string {
	return "AccountWithdrawn"
}

// Handle processa o evento de saque
func (h *AccountWithdrawnHandler) Handle(ctx context.Context, eventData []byte) error {
	var event account.AccountWithdrawnEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	log.Printf("Processando saque - Conta: %s, Valor: %.2f, Saldo Atual: %.2f",
		event.AccountID, event.Amount, event.CurrentBalance)

	// Exemplos de processamento:

	// 1. Verificar padrões de saque suspeitos
	// if h.isUnusualWithdrawal(event.AccountID, event.Amount) {
	//     h.securityService.FlagSuspiciousActivity(event.AccountID, event.Amount)
	// }

	// 2. Notificar sobre saldo baixo
	if event.CurrentBalance < 100 {
		log.Printf("AVISO: Saldo baixo na conta %s: %.2f",
			event.AccountID, event.CurrentBalance)
		// h.notificationService.SendLowBalanceAlert(event.AccountID, event.CurrentBalance)
	}

	// 3. Atualizar limites diários
	// h.limitService.UpdateDailyWithdrawal(event.AccountID, event.Amount)

	// 4. Enviar comprovante por email
	// h.emailService.SendWithdrawalReceipt(event.AccountID, event.Amount, event.CurrentBalance)

	// 5. Atualizar estatísticas de uso
	// h.analyticsService.TrackWithdrawal(event.AccountID, event.Amount, time.Now())

	log.Printf("Saque da conta %s processado com sucesso", event.AccountID)
	return nil
}
