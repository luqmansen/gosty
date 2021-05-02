package rabbitmq

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type rabbitRepo struct {
	uri  string
	conn *amqp.Connection
	channel
}

type channel struct {
	//reuse the channel so it doesn't exhaust the server
	// Also since on channel can't be used concurrently for different purposes
	//(declare, publish, etc), here I multiple channel for different action
	// reference: https://github.com/streadway/amqp/issues/170
	setQosChan       *amqp.Channel
	queueDeclareChan *amqp.Channel
	consumerChan     *amqp.Channel
	publisherChan    *amqp.Channel
}

func initChannel(connection *amqp.Connection, name string) *amqp.Channel {
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Errorf("trying to reconnect: %s", err.Error())
		initChannel(connection, name)
	}()
	rmqChannel, err := connection.Channel()
	if err != nil {
		log.Errorf("Failed to init %s channel: %s", name, err)
	}
	if rmqChannel != nil {
		rmqChannel.NotifyClose(c)
	}
	return rmqChannel
}

func NewRepository(uri string, client *amqp.Connection) repositories.Messenger {
	log.Debugf("Rabbitmq uri: %s", uri)

	return &rabbitRepo{
		uri:  uri,
		conn: client,
		channel: channel{
			setQosChan:       initChannel(client, "setQos"),
			queueDeclareChan: initChannel(client, "queueDeclare"),
			consumerChan:     initChannel(client, "consumer"),
			publisherChan:    initChannel(client, "publisher"),
		},
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
		log.Fatalf("cannot connect to rabbitmq: %s", err.Error())
	}
	if conn != nil {
		conn.NotifyClose(c)
	}

	return conn
}

func (r *rabbitRepo) Publish(data interface{}, queueName string) (err error) {

	q, err := r.queueDeclareChan.QueueDeclare(
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

	err = r.publisherChan.Publish(
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

	//declare queue name, in  case the queue haven't created
	q, err := r.queueDeclareChan.QueueDeclare(
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

	err = r.setQosChan.Qos(1, 0, true)
	if err != nil {
		log.Errorf("Failed to set QoS for the channel: %s", err)
	}

	msg, err := r.consumerChan.Consume(
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
