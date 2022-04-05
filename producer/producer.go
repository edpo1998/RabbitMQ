package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// Estructura del Objeto a recibir

type Game struct {
	Game_id int64 `json:"game_id"`
	Players int64 `json:"players"`
}

// ToJSON to be used for marshalling of Book type
func (g Game) ToJSON() []byte {
	ToJSON, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	return ToJSON
}

func main() {
	fmt.Println("Starting RabbitMQ producer...")
	time.Sleep(7 * time.Second)

	// Dial connection to broker
	conn, err := amqp.Dial(brokerAddr())
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Open a channel for communication
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare the Queue
	q, err := ch.QueueDeclare(
		queue(), // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgCount := 0

	// Get signal for finish
	doneCh := make(chan struct{})

	go func() {
		for {
			msgCount++
			game := new(Game)
			game.Game_id = int64(msgCount)
			game.Players = 10

			body := game.ToJSON()

			err = ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			log.Printf(" [x] Sent %s", body)
			failOnError(err, "Failed to publish a message")

			time.Sleep(5 * time.Second)
		}
	}()

	<-doneCh
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
