package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-bedrock/postgresx"
	"github.com/leetun2k2/go-bedrock/redisx"
	"github.com/leetun2k2/go-bedrock/restapix"
	"github.com/leetun2k2/go-bedrock/serverx"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

//go:embed default.yaml
var defaultConfig []byte

// RedisConfig holds application-level Redis configuration.
// Redis is disabled by default; when enabled, Config specifies client parameters.
type RedisConfig struct {
	Enabled       bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	redisx.Config `json:"config" mapstructure:"config" yaml:"config"`
}

// Config holds the complete runtime configuration.
//
// Environment variable overrides use double-underscore separators:
//
//	POSTGRES__URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
//	POSTGRES__MAX_CONNS=10
//	POSTGRES__MIN_CONNS=2
//	REDIS__ENABLED=true
//	REDIS__CONFIG__ADDR=127.0.0.1:6379
type Config struct {
	Logger     *logx.LoggerConfig `json:"logger" mapstructure:"logger" yaml:"logger"`
	Server     *serverx.Config    `json:"server" mapstructure:"server" yaml:"server"`
	RestfulApi *restapix.Config   `json:"restful_api" mapstructure:"restful_api" yaml:"restful_api"`
	Postgres   *postgresx.Config  `json:"postgres" mapstructure:"postgres" yaml:"postgres"`
	Redis      *RedisConfig       `json:"redis" mapstructure:"redis" yaml:"redis"`
}

func Load() *Config {
	_ = gotenv.Load(".env")

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBuffer(defaultConfig)); err != nil {
		panic(fmt.Sprintf("Failed to read viper config: %v", err))
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	cfg := &Config{
		Logger:     &logx.LoggerConfig{},
		Server:     &serverx.Config{},
		RestfulApi: &restapix.Config{},
		Postgres:   &postgresx.Config{},
		Redis:      &RedisConfig{},
	}
	if err := v.Unmarshal(cfg); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal config: %v", err))
	}

	return cfg
}
