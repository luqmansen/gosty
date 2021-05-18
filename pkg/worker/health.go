package worker

import (
	"fmt"
	hc "github.com/heptiolabs/healthcheck"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	util2 "github.com/luqmansen/gosty/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"net/http"
	"os"
	"time"
)

func InitHealthCheck(cfg *config.Configuration, connection *amqp.Connection, rabbitmqUri string) {
	health := hc.NewHandler()
	//too many goroutine might be a sign of a resource leak
	health.AddLivenessCheck("goroutine-threshold", hc.GoroutineCountCheck(200000))

	health.AddReadinessCheck("rabbitmq", util2.RabbitPingCheck(connection, rabbitmqUri))
	health.AddReadinessCheck("file-server", hc.HTTPGetCheck(cfg.FileServer.GetFileServerUri(), 30*time.Second))

	port := "8087"
	log.Infof("healthcheck running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), health); err != nil {
		log.Error(err)
	}
}
