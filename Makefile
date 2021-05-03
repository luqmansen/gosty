GOPATH:=$(shell go env GOPATH)
TAG:=$(shell git rev-parse --short=5 HEAD)

.PHONY: dev
api:
	FILE_MIN_SIZE_MB=50 nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run cmd/fileserver/main.go --signal SIGTERM

api-bin:
	CGO_ENABLED=0 go build -o build/apiserver/apiserver cmd/apiserver/main.go
worker-bin:
	CGO_ENABLED=0 go build -o build/worker/worker cmd/worker/main.go
fs-bin:
	CGO_ENABLED=0 go build -o build/fileserver/fileserver cmd/fileserver/main.go

all-bin: api-bin worker-bin fs-bin

cleanup:
	rm -rf build/*

docker-base-worker: cleanup
	docker build -t luqmansen/alpine-ffmpeg-mp4box -f docker/Dockerfile-alpine-ffmpeg-mp4box .

docker-web-dev:
	npm run build --prefix=web/
	docker build -t localhost:5000/gosty-web-dev -f docker/web.dev.Dockerfile .
	docker push localhost:5000/gosty-web-dev

docker-web:
	DOCKER_BUILDKIT=1 docker build -t luqmansen/gosty-web -f docker/web.Dockerfile .

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

push-all: docker-api docker-fs docker-worker docker-web

rollout-restart:
	# sometimes rabbitmq randomly wont start
	kubectl rollout restart statefulset -n gosty rabbit-rabbitmq
   	# sometimes i forgot to repush all the stuff to local registry
   	# happens if you use local registry and frequently need to stop the minikube
	kubectl rollout restart -f k8s/gosty/gosty-apiserver.yaml
	kubectl rollout restart -f k8s/gosty/gosty-worker.yaml
	kubectl rollout restart -f k8s/gosty/gosty-web.yaml