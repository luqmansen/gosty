package rabbitmq

import (
	"encoding/json"
	"github.com/cenkalti/backoff/v4"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

type rabbitRepo struct {
	conn *amqp.Connection
	uri  string
	//queue string
}

func NewRepository(uri string) repositories.Messenger {
	//TODO defer close connection somewhere
	log.Debugf("Rabbitmq uri: %s", uri)

	return &rabbitRepo{
		conn: connectWithBackoff(uri),
		uri:  uri,
	}
}
func connectWithBackoff(connectionUri string) (conn *amqp.Connection) {

	dial := func() (err error) {
		conn, err = amqp.Dial(connectionUri)
		if err != nil {
			log.Errorf("Failed to dial URI %s: %s, retrying.....", connectionUri, err)
			return
		}
		return
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	maxRetries := uint64(10)

	err := backoff.Retry(dial, backoff.WithMaxRetries(b, maxRetries))
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %s", err)
		return nil
	}
	return conn
}

func (r *rabbitRepo) Publish(data interface{}, queueName string) (err error) {

	if r.conn.IsClosed() {
		log.Debug("Connection is closed, attempting to open it")
		r.conn = connectWithBackoff(r.uri)
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
	} else {
		//log.Debugf("Success publish %s to %s queue", dataToSend, q.Name)
	}

	return err
}

func (r *rabbitRepo) ReadMessage(res chan<- interface{}, queueName string) {

	if r.conn.IsClosed() {
		r.conn = connectWithBackoff(r.uri)
	}

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
