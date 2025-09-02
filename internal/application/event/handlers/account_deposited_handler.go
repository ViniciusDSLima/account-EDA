package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/viniciuslima/account-EDA/internal/domain/account"
)

// AccountDepositedHandler processa eventos de depósito em conta
type AccountDepositedHandler struct {
	// Dependências como serviços de notificação, análise de fraude, etc.
}

// NewAccountDepositedHandler cria um novo handler para eventos de depósito
func NewAccountDepositedHandler() *AccountDepositedHandler {
	return &AccountDepositedHandler{}
}

// EventType retorna o tipo de evento que este handler processa
func (h *AccountDepositedHandler) EventType() string {
	return "AccountDeposited"
}

// Handle processa o evento de depósito
func (h *AccountDepositedHandler) Handle(ctx context.Context, eventData []byte) error {
	var event account.AccountDepositedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	log.Printf("Processando depósito - Conta: %s, Valor: %.2f, Saldo Atual: %.2f",
		event.AccountID, event.Amount, event.CurrentBalance)

	// Exemplos de processamento:

	// 1. Verificar depósitos suspeitos (anti-fraude)
	if event.Amount > 10000 {
		log.Printf("ALERTA: Depósito alto detectado na conta %s: %.2f",
			event.AccountID, event.Amount)
		// h.fraudService.AnalyzeDeposit(event.AccountID, event.Amount)
	}

	// 2. Enviar notificação push/SMS
	// h.notificationService.SendDepositNotification(event.AccountID, event.Amount)

	// 3. Atualizar relatórios em tempo real
	// h.reportingService.UpdateDailyDeposits(event.Amount)

	// 4. Verificar metas de poupança
	// if event.CurrentBalance >= goalAmount {
	//     h.goalService.NotifyGoalReached(event.AccountID)
	// }

	// 5. Atualizar sistema de cashback/rewards
	// h.rewardsService.ProcessDepositReward(event.AccountID, event.Amount)

	log.Printf("Depósito na conta %s processado com sucesso", event.AccountID)
	return nil
}
