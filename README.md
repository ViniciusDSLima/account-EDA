# Account-EDA

Aplicação de gerenciamento de contas utilizando arquitetura orientada a eventos (EDA) e padrão CQRS (Command Query Responsibility Segregation).

## Arquitetura

A aplicação segue os princípios de:

- **Domain-Driven Design (DDD)**: Organização do código centrada no domínio
- **Command Query Responsibility Segregation (CQRS)**: Separação de comandos e consultas
- **Event-Driven Architecture (EDA)**: Comunicação baseada em eventos
- **Outbox Pattern**: Garantia de entrega confiável de eventos

### Estrutura do Projeto

```
├── cmd
│   └── api               # Ponto de entrada da aplicação
├── internal
│   ├── domain            # Entidades e regras de domínio
│   │   └── account       
│   ├── application       # Casos de uso da aplicação
│   │   ├── command       # Comandos (escritas)
│   │   ├── query         # Consultas (leituras)
│   │   └── event         # Eventos e publicadores
│   └── infrastructure    # Implementações técnicas
│       ├── persistence   # Repositórios para persistência
│       ├── kafka         # Implementação de mensageria
│       └── api           # Handlers e rotas da API
└── docker-compose.yml    # Configuração dos serviços
```

## Requisitos

- Go 1.22+
- Docker e Docker Compose

## Como Executar

1. Clone o repositório
2. Execute os serviços com Docker Compose:

```bash
docker-compose up -d
```

3. Execute a aplicação:

```bash
go run cmd/api/main.go
```

## API REST

A API disponibiliza os seguintes endpoints:

### Contas

- `POST /accounts` - Criar uma conta
- `GET /accounts` - Listar todas as contas
- `GET /accounts/{id}` - Obter detalhes de uma conta
- `POST /accounts/{id}/deposit` - Realizar um depósito
- `POST /accounts/{id}/withdraw` - Realizar um saque

### Exemplo de Uso

Criar uma conta:

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"name":"João Silva","email":"joao@example.com"}'
```

Realizar um depósito:

```bash
curl -X POST http://localhost:8080/accounts/{id}/deposit \
  -H "Content-Type: application/json" \
  -d '{"amount":100.00}'
```

## Implementação CQRS

A aplicação implementa CQRS através da separação clara entre:

- **Commands**: Operações que modificam o estado (CreateAccount, Deposit, Withdraw)
- **Queries**: Operações que leem o estado (GetAccount, GetAccounts)
- **Events**: Notificações de mudanças de estado (AccountCreated, AccountDeposited)

## Padrão Outbox

Para garantir a entrega confiável de eventos, a aplicação utiliza o padrão Outbox:

1. **Persistência**: Eventos são salvos na tabela `outbox_events` antes da publicação
2. **Publicação**: Sistema tenta publicar eventos no Kafka imediatamente
3. **Processamento**: Worker de background processa eventos pendentes do outbox
4. **Dead Letter Queue**: Eventos que falharam são enviados para DLQ
5. **Retry**: Sistema tenta reprocessar eventos falhados automaticamente

Esta abordagem garante que nenhum evento seja perdido, mesmo em caso de falhas temporárias do Kafka.

## Worker de Eventos

A aplicação inclui um worker dedicado para processar eventos:

```bash
# Executar o worker
go run cmd/worker/main.go

# Executar múltiplas instâncias para escalabilidade
go run cmd/worker/main.go &
go run cmd/worker/main.go &
```

O worker consome eventos do Kafka e executa ações específicas baseadas no tipo de evento.

Esta separação permite otimizar cada caminho independentemente e escalar de acordo com as necessidades.