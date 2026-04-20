package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger/zap"
	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/grpc"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/rest"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logger zap.Config   `yaml:"logger"`
	GRPC   grpc.Config  `yaml:"grpc"`
	REST   rest.Config  `yaml:"rest"`
	Store  store.Config `yaml:"store"`
	JWT    jwt.Config   `yaml:"jwt"`
}

func MustLoad() Config {
	var cfg Config
	configPath := ParseFlag(cfg)

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		errMsg := fmt.Sprintf("Failed to read config on path(%s): %s", configPath, err)
		panic(errMsg)
	}

	accessSecret := os.Getenv("ACCESS_SECRET")
	refreshSecret := os.Getenv("REFRESH_SECRET")

	cfg.JWT.AccessSecret = accessSecret
	cfg.JWT.RefreshSecret = refreshSecret

	if err := validator.V.Struct(&cfg); err != nil {
		errMsg := fmt.Sprintf("Failed to validate config: %s", err)
		panic(errMsg)
	}

	return cfg
}

func ParseFlag(cfg Config) string {
	configPath := flag.String("config", "./configs/config.yaml", "config file path")
	configHelp := flag.Bool("help", false, "show configuration help")

	flag.Parse()

	if *configHelp {
		headerText := "Configuration options:"
		help, err := cleanenv.GetDescription(&cfg, &headerText)
		if err != nil {
			errMsg := fmt.Sprintf("error getting configuration description: %s", err)
			panic(errMsg)
		}
		fmt.Println(help)
		os.Exit(0)
	}

	return *configPath
}
