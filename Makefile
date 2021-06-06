GOOS=linux
GOPATH:=$(shell go env GOPATH)
CPUCOUNT:=$(shell grep -c ^processor /proc/cpuinfo)
GIT_COMMIT := $(shell git rev-list -1 HEAD)

.PHONY: dev
api:
	nodemon --exec go run cmd/apiserver/main.go --signal SIGTERM

wrk:
	nodemon --exec go run cmd/worker/main.go --signal SIGTERM

fs:
	nodemon --exec go run cmd/fileserver/main.go --signal SIGTERM

test:
	go clean -testcache
	go test ./pkg/apiserver/... -v --parallel $(CPUCOUNT)

web-dev:
	yarn --cwd ./web/ start

api-bin:
	rm -rf build/apiserver
	GOOS=$(GOOS) CGO_ENABLED=0 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" \
		-o build/apiserver/apiserver cmd/apiserver/main.go
worker-bin:
	rm -rf build/worker
	GOOS=$(GOOS) CGO_ENABLED=0 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" \
 		-o build/worker/worker cmd/worker/main.go
fs-bin:
	rm -rf build/fileserver
	GOOS=$(GOOS) CGO_ENABLED=0 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" \
 		-o build/fileserver/fileserver cmd/fileserver/main.go

fs-sync: fs-bin
	docker build -t luqmansen/gosty-fileserver -f docker/fileserver.Dockerfile .
	docker container rm fileserver-0 fileserver-1 fileserver-2 --force
	docker-compose up -d fileserver-0 fileserver-1 fileserver-2

all-bin: api-bin worker-bin fs-bin

run:
	docker-compose up

stop:
	docker-compose -f docker-compose.yaml down

docker-base-worker:
	docker build -t luqmansen/alpine-ffmpeg-mp4box -f docker/Dockerfile-alpine-ffmpeg-mp4box .

docker-web:
	DOCKER_BUILDKIT=1 docker build -t luqmansen/gosty-web -f docker/web.Dockerfile .
	docker push luqmansen/gosty-web


docker-web-local:
	yarn --cwd ./web/ build
	docker build -t localhost:5000/gosty-web-dev -f docker/web.dev.Dockerfile .
	docker push localhost:5000/gosty-web-dev
	echo "y" | docker-compose rm -s web
	docker-compose up -d web

docker-worker: worker-bin
	docker build -t luqmansen/gosty-worker -f docker/worker.Dockerfile .
	docker push luqmansen/gosty-worker

docker-worker-local: worker-bin
	docker build -t localhost:5000/gosty-worker -f docker/worker.Dockerfile .
	docker push localhost:5000/gosty-worker
	echo "y" | docker-compose rm -s worker
	docker-compose up --scale worker=1 -d worker

docker-fs:  fs-bin
	docker build -t luqmansen/gosty-fileserver -f docker/fileserver.Dockerfile .
	docker push luqmansen/gosty-fileserver

docker-fs-local:  fs-bin
	docker build -t localhost:5000/gosty-fileserver -f docker/fileserver.Dockerfile .
	docker push localhost:5000/gosty-fileserver
	echo "y" | docker-compose rm -s fileserver
	docker-compose up -d fileserver


docker-api:  api-bin
	docker build -t luqmansen/gosty-apiserver -f docker/apiserver.Dockerfile .
	docker build -t luqmansen/gosty-apiserver -f docker/apiserver.Dockerfile .

docker-api-local:  api-bin
	docker build -t localhost:5000/gosty-apiserver -f docker/apiserver.Dockerfile .
	docker push localhost:5000/gosty-apiserver
	echo "y" | docker-compose rm -s apiserver
	docker-compose up -d apiserver


push-all: docker-api docker-fs docker-worker docker-web
push-all-local: docker-api-local docker-fs-local docker-worker-local docker-web-local

generate-mock:
	mockgen --destination=mock/pkg/apiserver/repositories/rabbitmq/mock_rabbit.go --package mock_rabbitmq --source=pkg/apiserver/repositories/messaging.go

rollout-restart:
	# sometimes rabbitmq randomly wont start
	kubectl rollout restart statefulset -n gosty rabbit-rabbitmq
   	# sometimes i forgot to repush all the stuff to local registry
   	# happens if you use local registry and frequently need to stop the minikube
	kubectl rollout restart -f k8s/gosty/gosty-apiserver.yaml
	kubectl rollout restart -f k8s/gosty/gosty-worker.yaml
	kubectl rollout restart -f k8s/gosty/gosty-web.yaml