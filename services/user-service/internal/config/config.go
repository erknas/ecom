package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string           `yaml:"env"`
	GRPC        GRPCServerConfig `yaml:"grpc"`
	DatabaseURL string           `yaml:"database_url"`
	Postgres    PostgresConfig   `yaml:"postgres"`
}

type GRPCServerConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type PostgresConfig struct {
	MaxConns        int32         `yaml:"max_conns"`
	MinConns        int32         `yaml:"min_conns"`
	MaxConnLifeTime time.Duration `yaml:"max_conn_life_time"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
	CheckPeriod     time.Duration `yaml:"check_period"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist %s", path))
	}

	cfg := new(Config)

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic(err)
	}

	return cfg
}

func fetchConfigPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to the config")
	flag.Parse()

	return path
}
