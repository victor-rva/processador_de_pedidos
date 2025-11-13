# Processador de Pedidos

Um serviço de backend construído em Go para gerenciamento e processamento de pedidos, utilizando uma arquitetura limpa (Clean Architecture) e processamento assíncRono com RabbitMQ.

## Arquitetura

O sistema é projetado para ser um serviço orientado a eventos, desacoplando o recebimento do pedido do seu processamento.

1.  **API Service (`cmd/api`):** Um servidor HTTP que recebe novos pedidos. (Ver Análise abaixo).
2.  **RabbitMQ:** Atua como o "message broker", enfileirando os pedidos para garantir que não sejam perdidos.
3.  **Order Service (`cmd/order`):** Um "worker" que consome as mensagens da fila do RabbitMQ. Ele executa a regra de negócio (`usecase.CalculateFinalPrice`) e persiste o resultado final no banco de dados SQLite.

---

### Análise da Arquitetura Atual

Com base nos códigos fornecidos, o sistema está configurado da seguinte forma:

* **`cmd/order` (Consumidor):** Está **100% funcional**. Ele se conecta ao RabbitMQ, escuta a fila, recebe o JSON, chama o *use case* `CalculateFinalPrice` e salva o pedido no `db.sqlite3`.

* **`cmd/api` (Produtor):** Está funcionando como um **stub/mock**. O endpoint `GET /order` atual não lê um *payload* e não publica no RabbitMQ. Ele apenas executa um cálculo em memória com dados fixos (`entity.NewOrder("1", 10, 1)`) e o retorna.

Para finalizar o fluxo, a API precisa ser modificada para **Publicar** a mensagem, em vez de processá-la. (Veja a seção "Próximos Passos").

---

## Estrutura do Projeto

* **`cmd/`**: Pontos de entrada da aplicação.
    * `api/main.go`: O servidor HTTP (atualmente um *stub*).
    * `order/main.go`: O consumidor de mensageria (Worker).
* **`internal/`**: O "core" da aplicação.
    * **`entity/`**: Entidades de domínio (`Order`) e suas regras de validação.
    * **`usecase/`**: Regras de negócio (`CalculateFinalPrice`) que orquestram as entidades.
    * **`infra/`**: Implementações de ferramentas externas (banco de dados, gateways).
        * `database/order_repository.go`: Implementação do repositório de pedidos.
* **`pkg/`**: Pacotes reutilizáveis.
    * `rabbitmq/rabbitmq.go`: Funções utilitárias para conectar e interagir com o RabbitMQ.
* **`k8s/`**: Manifestos para deploy no Kubernetes.
* **`docker-compose.yaml`**: Define o serviço do RabbitMQ para desenvolvimento local.

---

## Como Executar o Fluxo Completo

Para testar o sistema da forma como ele foi *desenhado* (API -> RabbitMQ -> Worker -> DB), siga estes passos:

### 1. Inicie a Infraestrutura (RabbitMQ)

O `docker-compose.yaml` irá iniciar o RabbitMQ e expor a porta de gerenciamento.

```bash
docker-compose up -d