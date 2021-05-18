package repositories

type Messenger interface {
	ResourcesWatcher()
	Publish(data interface{}, queueName string) error
	ReadMessage(res chan<- interface{}, queueName string)
}
