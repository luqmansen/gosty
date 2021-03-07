GOPATH:=$(shell go env GOPATH)

.PHONY: dev
server:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run fileserver/main.go --signal SIGTERM

