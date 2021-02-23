package repositories

type MessageBrokerRepository interface {
	Publish(data interface{}, queueName string) error
	ReadMessage(res chan<- []byte, queueName string)
}
