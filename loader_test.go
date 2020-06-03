package zcfg

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	type (
		apiConfig struct {
			Port uint16 `json:"port" flag:"port" env:"API_PORT"`
		}

		clickhouseConfig struct {
			Host        string        `json:"host"        flag:"host"        env:"CLICKHOUSE_HOST"`
			User        string        `json:"user"        flag:"user"        env:"CLICKHOUSE_USER"`
			Database    string        `json:"database"    flag:"database"    env:"CLICKHOUSE_DATABASE"`
			ReadTimeout time.Duration `json:"readTimeout" flag:"readTimeout" env:"CLICKHOUSE_READ_TIMEOUT"`
		}

		kafkaConfig struct {
			Brokers       []string `yaml:"brokers"       flag:"brokers"       env:"KAFKA_BROKERS"`
			Topic         string   `yaml:"topic"         flag:"topic"         env:"KAFKA_TOPIC"`
			ConsumerGroup string   `yaml:"consumerGroup" flag:"consumerGroup" env:"KAFKA_CONSUMER_GROUP"`
		}

		redisPoolConfig struct {
			MaxActive    int           `json:"maxActive"    flag:"maxActive"    env:"REDIS_POOL_MAX_ACTIVE"`
			MaxIdle      int           `json:"maxIdle"      flag:"maxIdle"      env:"REDIS_POOL_MAX_IDLE"`
			IdleTimeout  time.Duration `json:"idleTimeout"  flag:"idleTimeout"  env:"REDIS_POOL_IDLE_TIMEOUT"`
			ConnLifetime time.Duration `json:"connLifetime" flag:"connLifetime" env:"REDIS_POOL_CONN_LIFETIME"`
		}

		redisConfig struct {
			Host         string           `json:"host"         flag:"host"         env:"REDIS_HOST"`
			Database     int              `json:"database"     flag:"database"     env:"REDIS_DATABASE"`
			ReadTimeout  time.Duration    `json:"readTimeout"  flag:"readTimeout"  env:"REDIS_READ_TIMEOUT"`
			WriteTimeout time.Duration    `json:"writeTimeout" flag:"writeTimeout" env:"REDIS_WRITE_TIMEOUT"`
			Pool         *redisPoolConfig `json:"pool"         flag:"pool"`
		}

		logConfig struct {
			Level string `json:"level" flag:"level" env:"LOG_LEVEL"`
		}

		config struct {
			API        *apiConfig        `json:"api"        flag:"api"`
			Clickhouse *clickhouseConfig `json:"clickhouse" flag:"clickhouse"`
			Kafka      *kafkaConfig      `json:"kafka"      flag:"kafka"`
			Redis      *redisConfig      `json:"redis"      flag:"redis"`
			Log        *logConfig        `json:"log"        flag:"log"`
			Debug      bool              `json:"debug"      flag:"debug"      env:"DEBUG"`
		}
	)

	cfgFile, err := ioutil.TempFile(os.TempDir(), "zcfg-test-config-*")
	require.Nil(t, err)

	defer os.Remove(cfgFile.Name())

	cfgFile.Write([]byte(
		`{
			"api": {
				"port": 8000
			},
			"clickhouse": {
				"host": "localhost:9000"
			},
			"kafka": {
				"brokers": [
					"localhost:9092"
				]
			},
			"redis": {
				"host": "localhost:6379"
			}
		}`,
	))

	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	var cfg config
	cfgLoader := New(&cfg, FromJSON(cfgFile.Name()), UseFlags(flagSet))

	unsetEnv := setTestEnv(testCaseEnv{
		"LOG_LEVEL": "INFO",
	})

	defer unsetEnv()

	flagSet.Parse([]string{
		"--api.port", "8001",
		"--clickhouse.user", "testuser",
		"--clickhouse.database", "testdb",
		"--clickhouse.readTimeout", "1s",
		"--kafka.brokers", "localhost:9092,localhost:9093",
		"--redis.database", "1",
		"--redis.pool.maxActive", "64",
		"--debug", "1",
	})

	expectedCfg := config{
		API: &apiConfig{
			Port: 8001,
		},
		Clickhouse: &clickhouseConfig{
			Host:        "localhost:9000",
			User:        "testuser",
			Database:    "testdb",
			ReadTimeout: 1 * time.Second,
		},
		Kafka: &kafkaConfig{
			Brokers: []string{"localhost:9092", "localhost:9093"},
		},
		Redis: &redisConfig{
			Host:     "localhost:6379",
			Database: 1,
			Pool: &redisPoolConfig{
				MaxActive: 64,
			},
		},
		Log: &logConfig{
			Level: "INFO",
		},
		Debug: true,
	}

	err = cfgLoader.Load()

	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

type testCaseEnv map[string]string

func setTestEnv(env testCaseEnv) (unsetFunc func()) {
	for k, v := range env {
		os.Setenv(k, v)
	}

	return func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}
}
