package zconf

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
			Host        string        `json:"host"     flag:"host"     env:"CLICKHOUSE_HOST"`
			User        string        `json:"user"     flag:"user"     env:"CLICKHOUSE_USER"`
			Database    string        `json:"database" flag:"database" env:"CLICKHOUSE_DATABASE"`
			ReadTimeout time.Duration `yaml:"readTimeout" flag:"readTimeout" env:"CLICKHOUSE_READ_TIMEOUT"`
		}

		kafkaConfig struct {
			Brokers       []string `yaml:"brokers"       flag:"brokers"       env:"KAFKA_BROKERS"`
			Topic         string   `yaml:"topic"         flag:"topic"         env:"KAFKA_TOPIC"`
			ConsumerGroup string   `yaml:"consumerGroup" flag:"consumerGroup" env:"KAFKA_CONSUMER_GROUP"`
		}

		logConfig struct {
			Level string `json:"level" flag:"level" env:"LOG_LEVEL"`
		}

		config struct {
			API        *apiConfig        `json:"api"        flag:"api"`
			Clickhouse *clickhouseConfig `json:"clickhouse" flag:"clickhouse"`
			Kafka      *kafkaConfig      `json:"kafka"      flag:"kafka"`
			Log        *logConfig        `json:"log"        flag:"log"`
			Debug      bool              `json:"debug"      flag:"debug"      env:"DEBUG"`
		}
	)

	cfgFile, err := ioutil.TempFile(os.TempDir(), "zconf-test-config-*")
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
			}
		}`,
	))

	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	var cfg config
	cfgLoader := New(&cfg).FromJSON(cfgFile.Name()).WithFlags(flagSet)

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
