package config

import (
	"flag"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env        string           `mapstructure:"env"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	HTTPServer HTTPServerConfig `mapstructure:"http_server"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
}

type HTTPServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`
	Issuer         string        `mapstructure:"issuer"`
}

type PostgresConfig struct {
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifeTime time.Duration `mapstructure:"max_conn_life_time"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
	CheckPeriod     time.Duration `mapstructure:"check_period"`
}

func MustLoad() *Config {
	configPath, envPath := fetchPaths()

	if configPath == "" {
		panic("config path is empty")
	}

	if envPath == "" {
		panic("env path is empty")
	}

	godotenv.Load(envPath)

	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	v.AutomaticEnv()

	v.BindEnv("http_server.addr", "ADDR")
	v.BindEnv("jwt.secret", "SECRET")
	v.BindEnv("postgres.user", "POSTGRES_USER")
	v.BindEnv("postgres.password", "POSTGRES_PASSWORD")
	v.BindEnv("postgres.database", "POSTGRES_DB")
	v.BindEnv("postgres.host", "POSTGRES_HOST")
	v.BindEnv("postgres.port", "POSTGRES_PORT")

	cfg := new(Config)

	if err := v.Unmarshal(cfg); err != nil {
		panic(err)
	}

	return cfg
}

func fetchPaths() (string, string) {
	var configPath, envPath string

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&envPath, "env", "", "path to env file")
	flag.Parse()

	return configPath, envPath
}
