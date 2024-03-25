package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

// Config
// TODO: env variable DB_PASSWORD
type Config struct {
	LogLevel   string `yaml:"log_level" env-default:"info"`
	HTTPServer `yaml:"http_server"`
	Storage    `yaml:"storage"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Storage struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"5432"`
	Database string `yaml:"database" env-default:"postgres"`
	Username string `yaml:"username" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
}

func GetConfig() *Config {
	pathToConfig := fetchConfigPath()
	if _, err := os.Stat(pathToConfig); os.IsNotExist(err) {
		log.Fatalf("file does not exist: %v", err)
	}
	var cfg Config

	if err := cleanenv.ReadConfig(pathToConfig, &cfg); err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	return &cfg
}

func fetchConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "config-path", "configs/config.yaml", "path to config file")
	flag.Parse()
	return configPath
}
