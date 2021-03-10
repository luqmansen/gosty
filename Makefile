GOPATH:=$(shell go env GOPATH)

.PHONY: dev
server:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run fileserver/main.go --signal SIGTERM

build-bin:
	CGO_ENABLED=0 go build -o build/worker/app cmd/worker/main.go
	CGO_ENABLED=0 go build -o build/fileserver/app fileserver/main.go

docker-base-worker:
	docker build -t luqmansen/alpine-ffmpeg-mp4box -f docker/Dockerfile-alpine-ffmpeg-mp4box .

docker-worker:
	CGO_ENABLED=0 go build -o build/worker/app cmd/worker/main.go
	docker build -t luqmansen/gosty-worker -f docker/Dockerfile-worker .

docker-fs:
	CGO_ENABLED=0 go build -o build/fileserver/app fileserver/main.go
	docker build -t luqmansen/gosty-fileserver -f docker/Dockerfile-fileserver .
