package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/viniciuslima/account-EDA/internal/application/event/handlers"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/kafka"
)

func main() {
	// Configuração do Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:29092"), ",")
	groupID := getEnv("CONSUMER_GROUP_ID", "account-events-worker")
	topic := getEnv("KAFKA_TOPIC", "account-events")

	// Criar consumidor
	consumer := kafka.NewEventConsumer(kafkaBrokers, groupID, topic)

	// Registrar handlers para cada tipo de evento
	consumer.RegisterHandler(handlers.NewAccountCreatedHandler())
	consumer.RegisterHandler(handlers.NewAccountDepositedHandler())
	consumer.RegisterHandler(handlers.NewAccountWithdrawnHandler())

	// Contexto para graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Canal para sinais do sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Iniciar consumidor em goroutine
	go func() {
		log.Println("Worker iniciado, consumindo eventos...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Erro no consumidor: %v", err)
		}
	}()

	// Aguardar sinal de interrupção
	<-sigChan
	log.Println("Recebido sinal de interrupção, encerrando worker...")

	// Cancelar contexto e parar consumidor
	cancel()
	if err := consumer.Stop(); err != nil {
		log.Printf("Erro ao parar consumidor: %v", err)
	}

	log.Println("Worker encerrado com sucesso")
}

// getEnv obtém uma variável de ambiente ou retorna um valor padrão
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
