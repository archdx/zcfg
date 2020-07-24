package main

import (
	"flag"
	"log"
	"time"

	"github.com/archdx/zcfg"

	"github.com/davecgh/go-spew/spew"
)

type Config struct {
	API        *APIConfig        `yaml:"api"        flag:"api"`
	Clickhouse *ClickhouseConfig `yaml:"clickhouse" flag:"clickhouse"`
	Kafka      *KafkaConfig      `yaml:"kafka"      flag:"kafka"`
	Log        *LogConfig        `yaml:"log"        flag:"log"`
}

type APIConfig struct {
	Port uint16 `yaml:"port" flag:"port" env:"API_PORT"`
}

type ClickhouseConfig struct {
	Host        string        `yaml:"host"        flag:"host"        env:"CLICKHOUSE_HOST"`
	User        string        `yaml:"user"        flag:"user"        env:"CLICKHOUSE_USER"`
	Database    string        `yaml:"database"    flag:"database"    env:"CLICKHOUSE_DATABASE"`
	ReadTimeout time.Duration `yaml:"readTimeout" flag:"readTimeout" env:"CLICKHOUSE_READ_TIMEOUT"`
}

type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"       flag:"brokers"       env:"KAFKA_BROKERS"`
	Topic         string   `yaml:"topic"         flag:"topic"         env:"KAFKA_TOPIC"`
	ConsumerGroup string   `yaml:"consumerGroup" flag:"consumerGroup" env:"KAFKA_CONSUMER_GROUP"`
}

type LogConfig struct {
	Level string `yaml:"level" flag:"level" env:"LOG_LEVEL"`
}

// LOG_LEVEL=DEBUG go run main.go -c config.yaml --api.port=8001 --clickhouse.readTimeout=1s --kafka.brokers=host1,host2
func main() {
	var cfg Config

	cfgLoader := zcfg.New(&cfg, zcfg.FromFile("config.yaml" /* default path */), zcfg.UseFlags(flag.CommandLine))
	// flag.CommandLine.Usage()
	if err := cfgLoader.Load(); err != nil {
		log.Fatal(err)
	}

	spew.Dump(cfg)
}
