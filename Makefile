GOPATH:=$(shell go env GOPATH)

.PHONY: dev
server:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run fileserver/main.go --signal SIGTERM

api-bin:
	CGO_ENABLED=0 go build -o build/apiserver/app cmd/apiserver/main.go
worker-bin:
	CGO_ENABLED=0 go build -o build/worker/app cmd/worker/main.go
fs-bin:
	CGO_ENABLED=0 go build -o build/fileserver/app fileserver/main.go

cleanup:
	rm -rf build/*

docker-base-worker: cleanup
	docker build -t luqmansen/alpine-ffmpeg-mp4box -f docker/Dockerfile-alpine-ffmpeg-mp4box .

docker-worker: cleanup worker-bin
	docker build -t luqmansen/gosty-worker -f docker/Dockerfile-worker .

docker-fs: cleanup fs-bin
	docker build -t luqmansen/gosty-fileserver -f docker/Dockerfile-fileserver .

docker-api: cleanup api-bin
	docker build -t luqmansen/gosty-apiserver -f docker/Dockerfile-apiserver .
