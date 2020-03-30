# zconf

### Description

Package zconf provides functionality to load application config with order of precedence:  
Command line arguments > Environment variable > Configuration file

The key difference from [viper](https://github.com/spf13/viper) is a ready to use config struct after load without need to additionally use any getters

### Example
```go
type Config struct {
	API        *APIConfig        `yaml:"api"        flag:"api"`
	Clickhouse *ClickhouseConfig `yaml:"clickhouse" flag:"clickhouse"`
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

type LogConfig struct {
	Level string `yaml:"level" flag:"level" env:"LOG_LEVEL"`
}

func main() {
	var cfg Config
	cfgLoader := zconf.New(&cfg).FromYAML("config.yaml").WithFlags(flag.CommandLine)

	err := cfgLoader.Load()
	if err != nil {
		log.Fatal(err)
	}
}
```
```bash
LOG_LEVEL=DEBUG go run main.go --api.port=8001 --clickhouse.user=testuser --clickhouse.database=testdb --clickhouse.readTimeout=1s
```

### TODO
- vault provider support
- more file types
- more examples
