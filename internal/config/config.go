package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Soloway struct {
	UserName string `yaml:"username" env:"SOLOWAY_USERNAME"`
	Password string `yaml:"password" env:"SOLOWAY_PASSWORD"`
}

type BQ struct {
	ServiceKeyPath string `yaml:"service_key_path"`
	ProjectID      string `yaml:"project_id"`
	DatasetID      string `yaml:"dataset_id"`
	TableID        string `yaml:"table_id"`
}

type TG struct {
	IsEnabled bool   `yaml:"is_enabled" env:"TG_ENABLED"`
	Token     string `yaml:"token" env:"TG_TOKEN"`
	Chat      int64  `yaml:"chat" env:"TG_CHAT"`
}

type GRPC struct {
	IP   string `yaml:"ip" env:"GRPC_IP"`
	Port int    `yaml:"port" env:"GRPC_PORT"`
}

type ServerConfig struct {
	TG      `yaml:"tg"`
	GRPC    `yaml:"grpc"`
	KeysDir string `yaml:"keys_dir" env:"KEYS_DIR"`
}

func NewServerConfig(filePath string, useEnv bool) (*ServerConfig, error) {
	cfg := &ServerConfig{}

	if useEnv {
		err := cleanenv.ReadEnv(cfg)
		if err != nil {
			return nil, fmt.Errorf("env error: %w", err)
		}
	} else {
		err := cleanenv.ReadConfig(filePath, cfg)
		if err != nil {
			return nil, fmt.Errorf("config file error: %w", err)
		}
	}

	return cfg, nil
}

type Report struct {
	ReportName       string `yaml:"report_name"`
	SpreadsheetID    string `yaml:"spreadsheet_id"`
	GoogleServiceKey string `yaml:"google_service_key"`
	ProjectID        string `yaml:"project_id"`
	DatasetID        string `yaml:"dataset_id"`
	Table            string `yaml:"table_id"`
	Days             int    `yaml:"period"`
}

func NewScheduleConfig(filePath string) (*ScheduleConfig, error) {
	cfg := &ScheduleConfig{}

	err := cleanenv.ReadConfig(filePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}

type ScheduleConfig struct {
	Time    string `yaml:"time"`
	GRPC    `yaml:"grpc"`
	BQ      `yaml:"bq"`
	Reports []Report `yaml:"reports"`
}
