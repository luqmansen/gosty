package rabbitmq

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type rabbitRepo struct {
	conn *amqp.Connection
	uri  string
	//queue string
}

func NewRepository(uri string) repositories.MessageBrokerRepository {
	//TODO defer close connection somewhere
	log.Debugf("Rabbitmq uri: %s", uri)
	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %s", err.Error())
	}

	return &rabbitRepo{
		conn: conn,
		uri:  uri,
	}
}

func (r rabbitRepo) Publish(data interface{}, queueName string) error {
	var err error
	if r.conn.IsClosed() {
		r.conn, err = amqp.Dial(r.uri)
		if err != nil {
			log.Fatalf("Failed to re-connect to rabbitmq: %s", err.Error())
		}
	}
	ch, err := r.conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}

	q, err := ch.QueueDeclare(
		queueName, // queueName
		true,      // durable
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
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         dataToSend,
		},
	)
	if err != nil {
		log.Errorf("Error when publish %s to %s queue: %s", dataToSend, q.Name, err)
	}

	return err
}

func (r rabbitRepo) ReadMessage(res chan<- interface{}, queueName string) {
	ch, err := r.conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}

	//declare queue name, in  case the queue haven't created
	q, err := ch.QueueDeclare(
		queueName, // queueName
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	msg, err := ch.Consume(
		q.Name,
		"",
		false,
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
		for d := range msg {
			res <- d
		}
	}()

	<-forever
}
