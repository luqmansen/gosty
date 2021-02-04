GOPATH:=$(shell go env GOPATH)

.PHONY: dev
dev:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

