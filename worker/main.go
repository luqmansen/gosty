package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("Worker started")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if (err != nil) || (ch == nil) {
		log.Fatal(err, ch)
	}
	q, err := ch.QueueDeclare(
		"task", // queueName
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	taskMsgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("Failed to register a consumer: %s", err)
	}
	forever := make(chan bool)

	//go func() {
	for d := range taskMsgs {
		processTaskSplit(d.Body)
	}
	//}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func processTaskSplit(d []byte) {

	var task models.Task
	err := json.Unmarshal(d, &task)
	if err != nil {
		print(err)
	}

	//download file
	filePath := fmt.Sprintf("./worker/tmp/%s", task.TaskSplit.Video.FileName)
	err = DownloadFile(filePath, "http://localhost:8001/files/" + task.TaskSplit.Video.FileName)
	if err != nil {
		log.Error(err)
	}

	log.Debugf("Processing task id:", task.Id.Hex())
	wd, _ := os.Getwd()
	dockerVol := fmt.Sprintf("%s:/work/tmp/", wd)
	cmd := exec.Command(
		"docker","run", "--rm", "-v", dockerVol,
		"sambaiz/mp4box", "-splits", fmt.Sprintf("%d", 1024 <<5),
		filePath)
	log.Debug(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = fmt.Sprintf("%s/worker/tmp",wd)
	fmt.Println(cmd.Dir)
	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	log.Info("Processing done: " + out.String())
	return
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
