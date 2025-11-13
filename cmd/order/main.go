package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victor-rva/processador_de_pedidos/internal/infra/database"
	"github.com/victor-rva/processador_de_pedidos/internal/usecase"
	"github.com/victor-rva/processador_de_pedidos/pkg/rabbitmq"
)

//import "honnef.co/go/tools/printf"

func main() {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close() //defer espera tudo rodar e depois execeuta o close (que encerra a conexão)
	orderRepository := database.NewOrderRepository(db)
	uc := usecase.NewCalculateFinalPrice(orderRepository)
	ch, err := rabbitmq.OpenChannel()
	if err != nil{
		panic(err)
	}
	defer ch.Close()
	msgRabbitmqChannel := make(chan amqp.Delivery) // criando um canal do tipo amqpDelivery
	go rabbitmq.Consume(ch, msgRabbitmqChannel) //escutando a fila // O go na frente faz com que seja uma thread 2
	rabbitmqWorker(msgRabbitmqChannel, uc) // thread 1
	// em uma thread ele le os dado do rabbitmq, os dados que recebe ele joga no canal msgRabbitmqChannel, o worker le o canal, pega os dados, executa o usecase e coloca no banco de dados.

	// input := usecase.OrderInput{
	// 	ID:    "1234",
	// 	Price: 10.0,
	// 	Tax:   1.0,
	// }
	// // usecase.Execute(input)
	// output, err := uc.Execute(input)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(output)
}

func rabbitmqWorker(msgChan chan amqp.Delivery, uc *usecase.CalculateFinalPrice){
	fmt.Println("Starting rabbitmq")
	for msg := range msgChan{
		var input usecase.OrderInput
		err := json.Unmarshal(msg.Body, &input)
		if err != nil{
			panic(err)
		}
		output, err := uc.Execute(input)
		if err != nil {
			panic(err)
		}
		msg.Ack(false)
		fmt.Println("Mensagem processada e salva no banco", output)
	}
}

