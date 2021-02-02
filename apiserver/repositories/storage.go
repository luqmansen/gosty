package repositories

import "bytes"

type StorageRepository interface {
	Add(file []bytes.Buffer) error
	Get(file string) []bytes.Buffer
	Delete(file string) error
}
