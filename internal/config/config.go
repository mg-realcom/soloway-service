package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
	"soloway/pkg/utils"
)

type Configuration struct {
	GRPC           GRPC      `yaml:"grpc" env-prefix:"GRPC_" env-required:"true"`
	Telemetry      Telemetry `yaml:"telemetry" env-prefix:"TELEMETRY_" env-required:"true"`
	Log            Log       `yaml:"log" env-prefix:"LOG_" env-required:"true"`
	AttachmentsDir string    `yaml:"attachments_dir" env:"ATTACHMENTS_DIR" env-required:"true"`
	KeysDir        string    `yaml:"keys_dir" env:"KEYS_DIR" env-required:"true"`
	PrometheusAddr string    `yaml:"prometheus_addr" env:"PROMETHEUS_ADDR" env-required:"true"`
}

type GRPC struct {
	Network string `yaml:"network" env:"NETWORK" env-required:"true"`
	Address string `yaml:"address" env:"ADDRESS" env-required:"true"`
}

type Telemetry struct {
	TracerName     string `yaml:"tracer_name" env:"TRACER_NAME" env-required:"true"`
	ServerName     string `yaml:"server_name" env:"SERVER_NAME" env-required:"true"`
	JaegerEndpoint string `yaml:"jaeger_endpoint" env:"JAEGER_ENDPOINT" env-required:"true"`
}

type Log struct {
	Level string `yaml:"level" env:"LEVEL" env-required:"true"`
}

func NewConfig() (*Configuration, error) {
	var envFiles []string

	if _, err := os.Stat(".env"); err == nil {
		log.Println("found .env file, adding it to env config files list")

		envFiles = append(envFiles, ".env")
	}

	cfg := &Configuration{}

	if len(envFiles) > 0 {
		for _, file := range envFiles {
			err := cleanenv.ReadConfig(file, cfg)
			if err != nil {
				return nil, errors.Wrapf(err, "error while read env config file: %s", err)
			}
		}
	} else {
		err := cleanenv.ReadEnv(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "error while opening env: %s", err)
		}
	}

	return cfg, nil
}

func (c *Configuration) Validation() error {
	if !utils.DirExists(c.KeysDir) {
		return errors.New("keys dir not found")
	}

	if !utils.DirExists(c.AttachmentsDir) {
		return errors.New("attachments dir not found")
	}

	return nil
}

type Soloway struct {
	UserName string `yaml:"username" env:"SOLOWAY_USERNAME"`
	Password string `yaml:"password" env:"SOLOWAY_PASSWORD"`
}
