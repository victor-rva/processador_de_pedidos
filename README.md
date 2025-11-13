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
````

  * **RabbitMQ Management:** `http://localhost:15672`
  * **Usuário:** `guest`
  * **Senha:** `guest`

### 2\. Inicie o Consumidor (Worker)

Em um terminal, inicie o `Order Service`. Ele ficará "escutando" por mensagens no RabbitMQ.

```bash
go run ./cmd/order/main.go
```

Você deverá ver a mensagem: `Starting rabbitmq`

### 3\. Publique um Pedido Manualmente (Teste)

Como a API ainda não está publicando mensagens, vamos fazer isso manualmente através da interface de gerenciamento do RabbitMQ.

1.  Acesse `http://localhost:15672`.
2.  Vá para a aba **Queues**.
3.  Clique na fila `orders` (ela deve ter sido criada pelo `cmd/order`).
4.  Abra a seção **Publish message**.
5.  Cole o seguinte JSON no campo **Payload**. Este é o DTO `OrderInput` esperado pelo seu *use case*:

<!-- end list -->

```json
{
    "id": "order-123",
    "price": 150.50,
    "tax": 15.0
}
```

6.  Clique em **Publish**.

### 4\. Verifique o Resultado

No terminal onde o `cmd/order/main.go` está rodando, você deverá ver a mensagem de confirmação do processamento:

```
Mensagem processada e salva no banco {order-123 150.5 15 165.5}
```

Isso confirma que seu *worker*, *use case* e *repositório* estão funcionando perfeitamente.

-----

## Próximos Passos (TO-DO)

Para finalizar o projeto, a única peça que falta é modificar o `cmd/api/main.go` para atuar como um produtor.

### 1\. Modificar o `OrderHandler`

O *handler* da API precisa:

1.  Ser um `POST /order` (em vez de `GET`).
2.  Ler o JSON do *body* da requisição.
3.  Publicar esse JSON no RabbitMQ.
4.  Retornar `HTTP 202 Accepted` (indicando que o pedido foi aceito para processamento).

**Exemplo de como o `OrderHandler` deveria se parecer:**

```go
// Em cmd/api/main.go

import (
    "encoding/json"
    "net/http"

    "[github.com/labstack/echo/v4](https://github.com/labstack/echo/v4)"
    "[github.com/victor-rva/projeto01_GO/internal/usecase](https://github.com/victor-rva/processador_de_pedidos/internal/usecase)" // Importar o DTO
    "[github.com/victor-rva/projeto01_GO/pkg/rabbitmq](https://github.com/victor-rva/processador_de_pedidos/pkg/rabbitmq)"     // Importar o pacote RabbitMQ
    // amqp "[github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)" // Pode ser necessário
)

// ... (conectar ao RabbitMQ no main) ...
// ch, err := rabbitmq.OpenChannel()
// if err != nil {
//     panic(err)
// }
// defer ch.Close()
// ...

// No main, injetar 'ch' no handler ou torná-lo acessível

func OrderHandler(c echo.Context) error {
    var input usecase.OrderInput

    // 1. Fazer o "bind" do JSON do body para o DTO
    if err := c.Bind(&input); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid JSON payload")
    }

    // 2. Serializar o DTO para JSON
    body, err := json.Marshal(input)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Error marshalling JSON")
    }

    // 3. Publicar no RabbitMQ
    // (Este passo assume que 'ch' está acessível)
    // Você precisará adaptar seu pkg/rabbitmq para ter uma função Publish
    /*
    err = rabbitmq.Publish(ch, body) // Exemplo de função a ser criada
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Error publishing to RabbitMQ")
    }
    */

    // 4. Retornar 202 Accepted
    return c.JSON(http.StatusAccepted, map[string]string{"status": "order received for processing"})
}
```

### 2\. Usar Variáveis de Ambiente

Não "hardcode" (fixar) conexões, nomes de fila ou portas. Use variáveis de ambiente para:

  * `RABBITMQ_URL`
  * `RABBITMQ_QUEUE`
  * `DB_SOURCE`
  * `API_PORT`

<!-- end list -->

```
```
