package repositories

type Messenger interface {
	ResourcesWatcher()
	Publish(data interface{}, queueName string) error
	ReadMessage(result chan<- interface{}, queueName string, setQos bool)
}
