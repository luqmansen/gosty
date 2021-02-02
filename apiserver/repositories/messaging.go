package repositories

type MessageBrokerRepository interface {
	Publish(topic string, msg interface{}) error
	Subscribe(topic string)
}
