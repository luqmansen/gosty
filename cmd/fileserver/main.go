package main

import (
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/fileserver"
	"os"
)

func main() {

	workDir, _ := os.Getwd()
	pathToServe := workDir + "/storage"

	port := util.GetEnv("PORT", "8001")
	host := util.GetEnv("FS_HOST", "0.0.0.0")

	fileServerHandler := fileserver.NewFileServerHandler(pathToServe)
	router := fileserver.NewRouter(fileServerHandler)
	server := fileserver.NewServer(port, host, router)

	server.Serve()
}
