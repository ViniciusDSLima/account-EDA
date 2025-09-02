# Worker de Processamento de Eventos

Este worker consome eventos do Kafka e executa ações baseadas neles.

## Arquitetura

O worker implementa o padrão de consumidor de eventos com:

1. **EventConsumer**: Responsável por consumir mensagens do Kafka
2. **EventHandlers**: Manipuladores específicos para cada tipo de evento
3. **Processamento assíncrono**: Cada evento é processado independentemente

## Como executar

```bash
# Executar o worker
go run cmd/worker/main.go

# Com variáveis de ambiente customizadas
KAFKA_BROKERS=localhost:29092 CONSUMER_GROUP_ID=my-worker go run cmd/worker/main.go
```

## Variáveis de Ambiente

- `KAFKA_BROKERS`: Lista de brokers do Kafka (padrão: localhost:29092)
- `CONSUMER_GROUP_ID`: ID do grupo de consumidores (padrão: account-events-worker)
- `KAFKA_TOPIC`: Tópico a ser consumido (padrão: account-events)

## Handlers Implementados

### AccountCreatedHandler
Processa eventos quando uma nova conta é criada. Pode ser usado para:
- Enviar email de boas-vindas
- Criar perfil em outros serviços
- Atualizar sistemas de analytics
- Notificar sistemas externos via webhook

### AccountDepositedHandler
Processa eventos de depósito. Pode ser usado para:
- Análise anti-fraude
- Notificações push/SMS
- Atualização de relatórios em tempo real
- Verificação de metas de poupança
- Sistema de rewards/cashback

### AccountWithdrawnHandler
Processa eventos de saque. Pode ser usado para:
- Detecção de atividades suspeitas
- Alertas de saldo baixo
- Controle de limites diários
- Envio de comprovantes
- Análise de padrões de uso

## Adicionando Novos Handlers

Para adicionar um novo handler:

1. Crie um novo arquivo em `internal/application/event/handlers/`
2. Implemente a interface `EventHandler`:
   ```go
   type EventHandler interface {
       Handle(ctx context.Context, event []byte) error
       EventType() string
   }
   ```
3. Registre o handler no worker:
   ```go
   consumer.RegisterHandler(handlers.NewMyEventHandler())
   ```

## Monitoramento

O worker fornece logs detalhados sobre:
- Eventos processados
- Erros de processamento
- Métricas de consumo

## Escalabilidade

Para escalar o processamento:

1. **Horizontal**: Execute múltiplas instâncias do worker com o mesmo `CONSUMER_GROUP_ID`
2. **Vertical**: Aumente os recursos da máquina
3. **Por tipo de evento**: Crie workers especializados para diferentes tipos de eventos

## Tratamento de Erros

- Erros de processamento não impedem o consumo de outras mensagens
- Mensagens com erro não são commitadas e serão reprocessadas
- Implementar dead letter queue para mensagens que falharam múltiplas vezes
