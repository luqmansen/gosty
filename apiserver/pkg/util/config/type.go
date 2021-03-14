package config

type Configuration struct {
	ApiServer     ApiServer     `mapstructure:",squash"`
	Database      Database      `mapstructure:",squash"`
	MessageBroker MessageBroker `mapstructure:",squash"`
	FileServer    FileServer    `mapstructure:",squash"`
}

type Database struct {
	Host     string `mapstructure:"MONGODB_SERVICE_HOST"`
	Port     string `mapstructure:"MONGODB_SERVICE_PORT"`
	Username string `mapstructure:"MONGODB_USERNAME"`
	Password string `mapstructure:"MONGODB_PASSWORD"`
	Name     string `mapstructure:"MONGODB_DATABASE"`
	Timeout  int    `mapstructure:"MONGODB_TIMEOUT"`
}

type ApiServer struct {
	Host string `mapstructure:"GOSTY_APISERVER_SERVICE_HOST"`
	Port string `mapstructure:"GOSTY_APISERVER_SERVICE_PORT"`
}

type FileServer struct {
	Host string `mapstructure:"GOSTY_FILESERVER_SERVICE_HOST"`
	Port string `mapstructure:"GOSTY_FILESERVER_SERVICE_PORT"`
}

type MessageBroker struct {
	Host     string `mapstructure:"RABBIT_RABBITMQ_SERVICE_HOST"`
	Port     string `mapstructure:"RABBIT_RABBITMQ_SERVICE_PORT"`
	Username string `mapstructure:"RABBITMQ_USERNAME"`
	Password string `mapstructure:"RABBITMQ_PASSWORD"`
}
