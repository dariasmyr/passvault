package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type GRPCConfig struct {
	Host         string        `yaml:"host" env-default:"localhost"`
	Port         int           `yaml:"port" env-default:"50051"`
	Timeout      time.Duration `yaml:"timeout" env-default:"5s"`
	RetriesCount int           `yaml:"reties_count" env-default:"60s"`
}

type Config struct {
	Env         string     `yaml:"env" env-default:"development"`
	GRPC        GRPCConfig `yaml:"grpc"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	Secret      string     `yaml:"secret" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error loading config file %s", err)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error loading config file %s", err)
	}

	return &cfg
}

// fetchConfigPath fetches domain path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)

	if !flag.Parsed() { // Check if flag has been parsed
		flag.StringVar(&res, "config", "", "path to config file")
	}
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	fmt.Println("Config path:", res)

	return res
}
