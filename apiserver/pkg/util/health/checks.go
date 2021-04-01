package health

import (
	"context"
	"fmt"
	hc "github.com/heptiolabs/healthcheck"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"time"
)

//Run this on goroutine, since http server will block main goroutine
func InitHealthCheck(cfg *config.Configuration) {
	health := hc.NewHandler()
	//too many goroutine might be a sign of a resource leak
	health.AddLivenessCheck("goroutine-threshold", hc.GoroutineCountCheck(200000))

	health.AddReadinessCheck("database", MongoDBPingCheck(cfg.Database.GetDatabaseUri(), 2*time.Second))
	health.AddReadinessCheck("rabbitmq", RabbitPingCheck(cfg.MessageBroker.GetMessageBrokerUri()))
	health.AddReadinessCheck("file-server", hc.HTTPGetCheck(cfg.FileServer.GetFileServerUri(), 2*time.Second))

	port := "8086"
	log.Infof("healthcheck running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), health); err != nil {
		log.Error(err)
	}
}

func MongoDBPingCheck(dbURI string, timeout time.Duration) hc.Check {
	return func() error {
		if dbURI == "" {
			return fmt.Errorf("database is nil")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		clientOptions := options.Client().ApplyURI(dbURI)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			return fmt.Errorf("failed to connect to db: %s", err)
		}

		if err = client.Ping(ctx, nil); err != nil {
			return fmt.Errorf("failed to ping db: %s", err)
		}
		return nil
	}
}

func RabbitPingCheck(rabbitURI string) hc.Check {
	return func() error {

		if rabbitURI == "" {
			return fmt.Errorf("rabbitmq URI is nil")
		}

		conn, err := amqp.Dial(rabbitURI)
		if err != nil {
			return fmt.Errorf("failed to connect to rabbitmq: %s", err)
		}

		ch, err := conn.Channel()
		if err != nil {
			return fmt.Errorf("failed to get rabbitmq channel: %s", err)
		}
		if ch == nil {
			return fmt.Errorf("rabbitmq channel channel empty: %s", err)
		}
		return nil
	}
}
