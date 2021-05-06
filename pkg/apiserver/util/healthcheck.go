package util

import (
	"context"
	"fmt"
	hc "github.com/heptiolabs/healthcheck"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"time"
)

//Run this on goroutine, since http server will block main goroutine
func InitHealthCheck(cfg *config.Configuration, mongoClient *mongo.Client, rabbitClient *amqp.Connection) {
	health := hc.NewHandler()
	//too many goroutine might be a sign of a resource leak
	health.AddLivenessCheck("goroutine-threshold", hc.GoroutineCountCheck(200000))

	health.AddReadinessCheck("database", MongoDBPingCheck(mongoClient, 10*time.Second))
	health.AddReadinessCheck("rabbitmq", RabbitPingCheck(rabbitClient))
	health.AddReadinessCheck("file-server", hc.HTTPGetCheck(cfg.FileServer.GetFileServerUri(), 30*time.Second))

	port := "8086"
	log.Infof("healthcheck running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), health); err != nil {
		log.Error(err)
	}
}

func MongoDBPingCheck(mongoClient *mongo.Client, timeout time.Duration) hc.Check {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := mongoClient.Ping(ctx, nil); err != nil {
			return fmt.Errorf("failed to ping db: %s", err)
		}

		return nil
	}
}

func RabbitPingCheck(connection *amqp.Connection) hc.Check {
	return func() error {
		// TODO:  try to reuse the channel instead of opening new channel everytime this endpoint got hit
		ch, err := connection.Channel()
		if err != nil {
			log.Errorf("failed to get rabbitmq channel: %s", err)
			return fmt.Errorf("failed to open rabbitmq channel: %s", err)
		}
		if ch == nil {
			log.Errorf("rabbitmq channel channel empty: %s", err)
			return fmt.Errorf("rabbitmq channel channel empty: %s", err)
		}

		q, err := ch.QueueDeclare("HEALTH_CHECK_QUEUE", false, true, false, false, nil)
		if err != nil {
			log.Errorf("failed create queue: %s", err)
			return fmt.Errorf("failed create queue: %s", err)
		}

		ch2, err := connection.Channel()
		if ch2 == nil {
			return fmt.Errorf("failed to open rabbitmq channel2: %s", err)
		}
		_, err = ch2.QueueDelete(q.Name, false, false, true)
		if err != nil {
			log.Errorf("failed delete queue: %s", err)
			return fmt.Errorf("failed delete queue: %s", err)
		}

		if err = ch.Close(); err != nil {
			log.Errorf("failed close channel: %s", err)
			return fmt.Errorf("failed close channel: %s", err)
		}
		if err = ch2.Close(); err != nil {
			log.Errorf("failed close channel2: %s", err)
			return fmt.Errorf("failed close channel2: %s", err)
		}

		return nil
	}
}
