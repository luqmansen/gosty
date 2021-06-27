package repositories

type Messenger interface {
	ResourcesWatcher() // TODO: remove this unused function, make sure to regenerate mock
	Publish(data interface{}, queueName string) error
	ReadMessage(result chan<- interface{}, queueName string, setQos bool)
}
