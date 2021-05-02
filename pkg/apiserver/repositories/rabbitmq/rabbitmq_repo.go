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
}

func NewRepository(uri string, client *amqp.Connection) repositories.Messenger {
	log.Debugf("Rabbitmq uri: %s", uri)

	return &rabbitRepo{
		conn: client,
		uri:  uri,
	}
}

func NewRabbitMQConn(connectionUri string) (conn *amqp.Connection) {
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Errorf("trying to reconnect: %s", err.Error())
		NewRabbitMQConn(connectionUri)
	}()

	conn, err := amqp.Dial(connectionUri)
	if err != nil {
		log.Fatalf("cannot connect: %s", err.Error())
	}
	if conn != nil {
		conn.NotifyClose(c)
	}

	return conn
}

func (r *rabbitRepo) Publish(data interface{}, queueName string) (err error) {

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
	} else {
		//log.Debugf("Success publish %s to %s queue", dataToSend, q.Name)
	}

	return err
}

func (r *rabbitRepo) ReadMessage(res chan<- interface{}, queueName string) {

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

	err = ch.Qos(1, 0, true)
	if err != nil {
		log.Errorf("Failed to set QoS for the channel: %s", err)
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
