FROM go-mp4box:latest
ARG SERVICE_NAME
ARG PORT

WORKDIR /app

ADD . .

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build ${SERVICE_NAME}/main.go" --command=./app