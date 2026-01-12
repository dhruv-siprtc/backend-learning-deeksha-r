package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

type ServerConf struct {
	Port            string `env:"HTTP_PORT" envDefault:":8080"`
	QueueLen        int    `env:"HTTP_QUEUE_LEN" envDefault:"0"`
	ErrorQueueLen   int    `env:"HTTP_ERROR_QUEUE_LEN" envDefault:"10000"`
	HttpMode        string `env:"HTTP_MODE" envDefault:"release"`
	ServiceLauncher string `env:"SERVICE_LAUNCHER" envDefault:"CONSUMER"`
}

type RedisConf struct {
	RedisHost     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisUser     string `env:"REDIS_USER"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       string `env:"REDIS_DB" envDefault:"0"`
}

type PostgresConf struct {
	User     string `env:"PGUSER" envDefault:"postgres"`
	Host     string `env:"PGHOST" envDefault:"localhost"`
	Port     string `env:"PGPORT" envDefault:"5432"`
	Password string `env:"PGPASSWORD" envDefault:"postgres"`
	DB       string `env:"PGDATABASE" envDefault:"user_management"`
}

type ConsumerConf struct {
	RabbitMQUrl    string `env:"CONSUMER_RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
	QueueName      string `env:"CONSUMER_QUEUE_NAME" envDefault:"user_task_queue"`
	ExchangeName   string `env:"CONSUMER_EXCHANGE_NAME" envDefault:"user_task_exchange"`
	BindingKeyName string `env:"CONSUMER_BINDING_KEY_NAME" envDefault:"user_task_binding_key"`
	DelayQueueName string `env:"CONSUMER_DELAY_QUEUE_NAME" envDefault:"user_task_delay_queue"`
	FailedQueue    string `env:"CONSUMER_FAILED_QUEUE" envDefault:"user_task_failed_queue"`
	TimeoutQueue   string `env:"CONSUMER_TIMEOUT_QUEUE" envDefault:"user_task_timeout_queue"`
	QueueTaskName  string `env:"CONSUMER_QUEUE_TASK_NAME" envDefault:"user_task"`
}

type LoggingConf struct {
	Level    string `env:"LOGGING_LEVEL" envDefault:"debug"`
	Facility string `env:"LOGGING_FACILITY" envDefault:"local1"`
	Tag      string `env:"LOGGING_TAG" envDefault:"taskrouter"`
	Format   string `env:"LOGGING_FORMAT" envDefault:"json"`
	Sentry   string `env:"LOGGING_SENTRY" envDefault:""`
	Syslog   string `env:"LOGGING_SYSLOG" envDefault:"127.0.0.1:514"`
}

type UserTaskProducer struct {
	RabbitMQUrl    string `env:"USER_TASK_RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
	QueueName      string `env:"USER_TASK_QUEUE_NAME" envDefault:"user_task_queue"`
	ExchangeName   string `env:"USER_TASK_EXCHANGE_NAME" envDefault:"user_task_exchange"`
	BindingKeyName string `env:"USER_TASK_BINDING_KEY_NAME" envDefault:"user_task_binding_key"`
	DelayQueueName string `env:"USER_TASK_DELAY_QUEUE_NAME" envDefault:"user_task_delay_queue"`
	FailedQueue    string `env:"USER_TASK_FAILED_QUEUE" envDefault:"user_task_failed_queue"`
	TimeoutQueue   string `env:"USER_TASK_TIMEOUT_QUEUE" envDefault:"user_task_timeout_queue"`
	QueueTaskName  string `env:"USER_TASK_QUEUE_TASK_NAME" envDefault:"user_task"`
}

type QueueMangerConfig struct {
	Logging          LoggingConf
	Server           ServerConf
	ConsumerConf     ConsumerConf
	Redis            RedisConf
	Postgres         PostgresConf
	UserTaskProducer UserTaskProducer
}

var Config QueueMangerConfig

func InitConfig() {
	if err := env.Parse(&Config); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
