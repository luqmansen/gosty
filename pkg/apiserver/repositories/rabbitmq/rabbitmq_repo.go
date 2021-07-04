package rabbitmq

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type rabbitRepo struct {
	uri  string
	conn *amqp.Connection
	channel
	connectionWatcherChan chan error
}

/*
Reuse the channel so it doesn't exhaust the server, since channel
can't be used concurrently for different purposes (declare, publish, etc),
here I multiple channel for different action.
reference: https://github.com/streadway/amqp/issues/170
*/
type channel struct {
	setQosChan       *amqp.Channel
	queueDeclareChan *amqp.Channel
	consumerChan     *amqp.Channel
	publisherChan    *amqp.Channel
}

func initRabbitMQChannel(connection *amqp.Connection, name string) *amqp.Channel {
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Errorf("trying to reconnect: %s", err.Error())
		initRabbitMQChannel(connection, name)
	}()
	if connection != nil {
		rmqChannel, err := connection.Channel()
		if err != nil {
			log.Errorf("Failed to init %s channel: %s", name, err)
		}
		if rmqChannel != nil {
			rmqChannel.NotifyClose(c)
		}
		return rmqChannel
	}
	return nil
}

func NewRepository(uri string, client *amqp.Connection) repositories.Messenger {
	log.Debugf("Rabbitmq uri: %s", uri)

	connWatcher := make(chan error)

	return &rabbitRepo{
		uri:  uri,
		conn: client,
		channel: channel{
			setQosChan:       initRabbitMQChannel(client, "setQos"),
			queueDeclareChan: initRabbitMQChannel(client, "queueDeclare"),
			consumerChan:     initRabbitMQChannel(client, "consumer"),
			publisherChan:    initRabbitMQChannel(client, "publisher"),
		},
		connectionWatcherChan: connWatcher,
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
		log.Errorf("cannot connect to rabbitmq: %s", err.Error())
		NewRabbitMQConn(connectionUri)
	}
	if conn != nil {
		conn.NotifyClose(c)
		return conn
	} else {
		log.Fatalf("Connection is nil, retrying")
		NewRabbitMQConn(connectionUri)
	}
	return conn
}

// ResourcesWatcher will be endpoint of all resource
// that need to be watched related to rabbitmq
func (r *rabbitRepo) ResourcesWatcher() {}

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

/*
ReadMessage parameter setQos need to be set to true for all read message on apiserver,
since it behaves differently from worker that need to read message 1-by-1, apiserver
can just consume all message then directly process them
*/
func (r *rabbitRepo) ReadMessage(result chan<- interface{}, queueName string, setQos bool) {

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
		log.Errorf("Failed to declare queue: %s", err)
	}
	//Consumer need to use same channel where prefetch count is set,
	//so that consumer will follow that prefetch rule.
	ch, err := r.conn.Channel()
	if ch != nil {
		if setQos {
			err = ch.Qos(viper.GetInt("TASK_PREFETCH_COUNT"), 0, true)
			defer func() {
				if err := ch.Close(); err != nil {
					log.Errorf("Failed to close channel on read message: %s", err)
				}
			}()
		}
	}

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

	go func() {
		for d := range msg {
			result <- d
		}
	}()

	select {}
}
