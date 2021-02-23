package repositories

import "github.com/luqmansen/gosty/apiserver/model"

type MessageBrokerRepository interface {
	PublishTask(task *model.Task) error
	ReadMessage(res chan<- []byte)
}
