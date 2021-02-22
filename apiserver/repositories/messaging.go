package repositories

type MessageBrokerRepository interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic string)
}
