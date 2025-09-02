package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/viniciuslima/account-EDA/internal/application/command"
	"github.com/viniciuslima/account-EDA/internal/application/query"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/api"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/kafka"
	"github.com/viniciuslima/account-EDA/internal/infrastructure/persistence"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "account")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err := persistence.RunMigrations(db); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:29092"), ",")
	eventPublisher := kafka.NewEventPublisher(kafkaBrokers)
	defer eventPublisher.Close()

	if err := eventPublisher.CreateTopics(); err != nil {
		log.Printf("Aviso: Não foi possível criar/verificar tópicos Kafka: %v", err)
		log.Printf("Os tópicos serão criados automaticamente quando o Kafka estiver disponível")
	}

	accountRepo, err := persistence.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	outboxRepo := persistence.NewOutboxRepository(db)

	outboxProcessor := kafka.NewOutboxProcessor(
		outboxRepo,
		eventPublisher,
		50,            // tamanho do lote
		5*time.Second, // intervalo de processamento
		5,             // máximo de tentativas
	)
	outboxProcessor.Start()
	defer outboxProcessor.Stop()

	createAccountHandler := command.NewCreateAccountHandler(accountRepo, eventPublisher, outboxRepo)
	depositHandler := command.NewDepositHandler(accountRepo, eventPublisher)
	withdrawHandler := command.NewWithdrawHandler(accountRepo, eventPublisher)

	accountQuery := query.NewAccountQueryHandler(accountRepo)

	accountHandler := api.NewAccountHandler(
		createAccountHandler,
		depositHandler,
		withdrawHandler,
		accountQuery,
	)

	e := api.SetupRoutes(accountHandler)

	port := getEnv("PORT", "8080")
	go func() {
		if err := e.Start(":" + port); err != nil {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	log.Printf("Servidor iniciado na porta %s", port)
	log.Printf("Processador de outbox iniciado com intervalo de %v", 5*time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Desligando o servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Servidor encerrado com sucesso")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
