package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Game struct {
	Game_id int64 `json:"game_id"`
	Players int64 `json:"players"`
}

func FromJSON(data []byte) Game {
	game := Game{}
	err := json.Unmarshal(data, &game)
	if err != nil {
		panic(err)
	}
	return game
}

func main() {
	fmt.Println("Starting RabbitMQ consumer...")
	time.Sleep(7 * time.Second)

	// Dial connection to broker
	conn, err := amqp.Dial(brokerAddr())
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	//Open communication channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	//Declare the queue
	q, err := ch.QueueDeclare(
		queue(), // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// register and fet a message consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %+v", FromJSON(d.Body))
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func brokerAddr() string {
	brokerAddr := os.Getenv("BROKER_ADDR")
	if len(brokerAddr) == 0 {
		brokerAddr = "amqp://guest:guest@localhost:5672/"
	}
	return brokerAddr
}

func queue() string {
	queue := os.Getenv("QUEUE")
	if len(queue) == 0 {
		queue = "default-queue"
	}
	return queue
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
