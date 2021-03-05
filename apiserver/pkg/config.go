package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strings"
)

type Configuration struct {
	Server   Server
	Database Database
}

type Database struct {
	URI      string
	Database string
	Timeout  int
}

type Server struct {
	Host string
	Port string
}

var config Configuration

func InitConfig() {
	viper.SetDefault("DEPLOY", "PROD")

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

	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal error failed to decode to struct: %s \n", err))
	}

}

func GetConfig() *Configuration {
	return &config
}
