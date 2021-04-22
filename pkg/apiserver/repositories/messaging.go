package repositories

type Messenger interface {
	Publish(data interface{}, queueName string) error
	ReadMessage(res chan<- interface{}, queueName string)
}
