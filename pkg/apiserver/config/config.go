package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

func LoadConfig(path string) *Configuration {

	if os.Getenv("DEPLOY") != "PROD" {
		log.SetLevel(log.DebugLevel)
		log.SetOutput(os.Stdout)
	}

	formatter := &prefixed.TextFormatter{
		ForceColors:     true,
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	formatter.SetColorScheme(&prefixed.ColorScheme{
		PrefixStyle:    "blue+b",
		TimestampStyle: "white+h",
	})
	log.SetFormatter(formatter)

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Error(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	conf := &Configuration{}
	err = viper.Unmarshal(conf)
	if err != nil {
		log.Error(fmt.Errorf("Fatal error failed to decode to struct: %s \n", err))
	}
	return conf
}

func (d Database) GetDatabaseUri() string {
	return fmt.Sprint(d.DbUri)
}

func (d Database) FormatGetDatabaseUri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		d.Username, d.Password, d.Host, d.Port, d.Name)
}

func (m MessageBroker) GetMessageBrokerUri() string {
	return fmt.Sprint(m.MbUri)
}

func (m MessageBroker) FormatGetMessageBrokerUri() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s",
		m.Username, m.Password, m.Host, m.Port)
}

func (f FileServer) GetFileServerUri() string {
	return fmt.Sprintf("http://%s:%s", f.Host, f.Port)
}
