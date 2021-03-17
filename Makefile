GOPATH:=$(shell go env GOPATH)
TAG:=$(shell git rev-parse --short=5 HEAD)

.PHONY: dev
server:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run cmd/fileserver/main.go --signal SIGTERM

api-bin:
	CGO_ENABLED=0 go build -o build/apiserver/app cmd/apiserver/main.go
worker-bin:
	CGO_ENABLED=0 go build -o build/worker/app cmd/worker/main.go
fs-bin:
	CGO_ENABLED=0 go build -o build/fileserver/app cmd/fileserver/main.go

cleanup:
	rm -rf build/*

docker-base-worker: cleanup
	docker build -t luqmansen/alpine-ffmpeg-mp4box -f docker/Dockerfile-alpine-ffmpeg-mp4box .

docker-web:
	#docker build -t luqmansen/gosty-worker -f docker/worker.Dockerfile .
	DOCKER_BUILDKIT=1 docker build -t localhost:5000/gosty-web -f docker/web.Dockerfile .
	docker push localhost:5000/gosty-web


docker-worker: cleanup worker-bin
	#docker build -t luqmansen/gosty-worker -f docker/worker.Dockerfile .
	docker build -t localhost:5000/gosty-worker -f docker/worker.Dockerfile .
	docker push localhost:5000/gosty-worker

docker-fs: cleanup fs-bin
	#docker build -t luqmansen/gosty-fileserver -f docker/fileserver.Dockerfile .
	docker build -t localhost:5000/gosty-fileserver -f docker/fileserver.Dockerfile .
	docker push localhost:5000/gosty-fileserver

docker-api: cleanup api-bin
	#docker build -t luqmansen/gosty-apiserver:$(TAG) -f docker/apiserver.Dockerfile .
	#docker build -t luqmansen/gosty-apiserver -f docker/apiserver.Dockerfile .

	#for using docker local registry
	docker build -t localhost:5000/gosty-apiserver -f docker/apiserver.Dockerfile .
	docker push localhost:5000/gosty-apiserver

push-all: docker-api docker-fs docker-worker

roll-api: docker-api
	kubectl rollout restart -f k8s/gosty-apiserver.yaml