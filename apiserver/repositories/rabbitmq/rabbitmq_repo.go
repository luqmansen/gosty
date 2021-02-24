package rabbitmq

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type rabbitRepo struct {
	conn *amqp.Connection
	//queue string
}

func NewRabbitMQRepo(uri string) repositories.MessageBrokerRepository {
	//TODO defer close connection somewhere
	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatal(err)
	}
	return &rabbitRepo{
		conn: conn,
	}
}

func (r rabbitRepo) Publish(data interface{}, queueName string) error {
	ch, err := r.conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}

	q, err := ch.QueueDeclare(
		queueName, // queueName
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	dataToSend, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        dataToSend,
		},
	)
	log.Debug("Success publish message")

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (r rabbitRepo) ReadMessage(res chan<- []byte, queueName string) {
	ch, err := r.conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	log.Errorf("Failed to register a consumer: %s", err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
