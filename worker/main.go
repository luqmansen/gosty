package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}

	taskMsgs, err := ch.Consume(
		"task",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("Failed to register a consumer: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range taskMsgs {
			fmt.Println("received message")
			fmt.Println(d.MessageId)
			fmt.Println(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
