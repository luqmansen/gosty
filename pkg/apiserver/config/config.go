package config

import (
	"fmt"
	"github.com/keepeye/logrus-filename"
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
	filenameHook := filename.NewHook()
	filenameHook.Field = "SRC"
	log.AddHook(filenameHook)

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
	log.Debugf("Split task minimum file: %s", viper.GetString("FILE_MIN_SIZE_MB"))
	return conf
}
func (d Database) GetDatabaseUri() string {
	if d.DbUri != "" {
		return fmt.Sprint(d.DbUri)
	}
	// TODO: remove this part, just simplify to single connection string
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		d.Username, d.Password, d.Host, d.Port, d.Name)
}

func (m MessageBroker) GetMessageBrokerUri() string {
	if m.MbUri != "" {
		return fmt.Sprint(m.MbUri)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s",
		m.Username, m.Password, m.Host, m.Port)
}

func (f FileServer) GetFileServerUri() string {
	return fmt.Sprintf("http://%s:%s", f.Host, f.Port)
}
