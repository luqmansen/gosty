package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/fileserver"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

var gitCommit string

func main() {
	workDir, _ := os.Getwd()
	folder := util.GetEnv("STORAGE", "storage")
	pathToServe := workDir + "/" + folder

	port := util.GetEnv("PORT", "8001")
	host := util.GetEnv("POD_IP", "0.0.0.0")
	address := fmt.Sprintf("%s:%s", host, port)
	peerHost := util.GetEnv("FILESERVER_PEER_HOST", "")
	peers := strings.Split(peerHost, ",")
	selfHost := os.Getenv("HOSTNAME")
	var peerLists []string
	for _, peer := range peers {
		if peer != "" && !strings.Contains(peer, selfHost) {
			peerLists = append(peerLists, fmt.Sprintf("%s:%s", peer, port))
		}
	}
	log.Infof("Peer list: %s", peerLists)
	fileServerHandler := fileserver.NewFileServerHandler(pathToServe, peerLists, address)
	router := fileserver.NewRouter(fileServerHandler)
	server := fileserver.NewServer(address, router)
	getVersion(router)

	go server.Serve()

	go fileServerHandler.InitialSync()
	go fileServerHandler.ExecuteSynchronization()

	forever := make(chan bool)
	<-forever
}

func getVersion(router *chi.Mux) {
	router.Get("/version", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(gitCommit))
	})
}
