package main

import (
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/fileserver"
	"os"
)

func main() {

	workDir, _ := os.Getwd()
	folder := util.GetEnv("STORAGE", "storage")
	pathToServe := workDir + "/" + folder

	port := util.GetEnv("PORT", "8001")
	host := util.GetEnv("FS_HOST", "0.0.0.0")
	peerHost := []string{"0.0.0.0:8001", "0.0.0.0:8002"}
	selfHost := fmt.Sprintf("%s:%s", host, port)
	var excludedSelfHost []string
	for _, h := range peerHost {
		if h != selfHost {
			excludedSelfHost = append(excludedSelfHost, h)
		}
	}

	fileServerHandler := fileserver.NewFileServerHandler(pathToServe, excludedSelfHost, selfHost)
	router := fileserver.NewRouter(fileServerHandler)
	server := fileserver.NewServer(selfHost, router)
	go fileServerHandler.ExecuteSynchronization()

	server.Serve()
}
