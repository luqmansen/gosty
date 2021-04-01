package worker

import (
	"fmt"
	hc "github.com/heptiolabs/healthcheck"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	util "github.com/luqmansen/gosty/apiserver/pkg/util/health"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

func InitHealthCheck(cfg *config.Configuration) {
	health := hc.NewHandler()
	//too many goroutine might be a sign of a resource leak
	health.AddLivenessCheck("goroutine-threshold", hc.GoroutineCountCheck(200000))

	health.AddReadinessCheck("rabbitmq", util.RabbitPingCheck(cfg.MessageBroker.GetMessageBrokerUri()))
	health.AddReadinessCheck("file-server", hc.HTTPGetCheck(cfg.FileServer.GetFileServerUri(), 2*time.Second))

	port := "8002"
	log.Infof("healthcheck running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), health); err != nil {
		log.Error(err)
	}
}
