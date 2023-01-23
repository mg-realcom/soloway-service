package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App     `yaml:"app"`
	Soloway `yaml:"soloway"`
	BQ      `yaml:"bq"`
	TG      `yaml:"tg"`
}

type App struct {
	Name       string `yaml:"name"`
	DeltaDay   int    `yaml:"delta_day"`
	DeltaMonth int    `yaml:"delta_month"`
	DeltaYear  int    `yaml:"delta_year"`
}

type Soloway struct {
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
}

type BQ struct {
	ServiceKeyPath string `yaml:"service_key_path"`
	ProjectID      string `yaml:"project_id"`
	DatasetID      string `yaml:"dataset_id"`
	TableID        string `yaml:"table_id"`
}

type TG struct {
	Token string `yaml:"token"`
	Chat  int64  `yaml:"chat"`
}

func NewConfig(filePath string) (*Config, error) {
	cfg := &Config{}
	fmt.Println(filePath)
	err := cleanenv.ReadConfig(filePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
