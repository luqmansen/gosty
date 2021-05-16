package main

import (
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/fileserver"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func main() {
	workDir, _ := os.Getwd()
	folder := util.GetEnv("STORAGE", "storage")
	pathToServe := workDir + "/" + folder

	port := util.GetEnv("PORT", "8001")
	host := util.GetEnv("POD_IP", "0.0.0.0")
	peerHost := util.GetEnv("FILESERVER_PEER_HOST", "")
	peers := strings.Split(peerHost, ",")
	selfHost := viper.GetString("HOSTNAME")
	var excludedSelfHost []string
	for _, peer := range peers {
		if !strings.Contains(peer, selfHost) && peer != "" {
			excludedSelfHost = append(excludedSelfHost, peer)
		}
	}

	fileServerHandler := fileserver.NewFileServerHandler(pathToServe,
		excludedSelfHost, fmt.Sprintf("%s:%s", host, port))
	router := fileserver.NewRouter(fileServerHandler)
	server := fileserver.NewServer(selfHost, router)
	go server.Serve()

	go fileServerHandler.InitialSync() // don't run this on goroutine, need to be finished first
	go fileServerHandler.ExecuteSynchronization()

	forever := make(chan bool)
	<-forever
}
