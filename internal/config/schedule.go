package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type ScheduleConfig struct {
	Destination string   `yaml:"destination" env-required:"true"`
	Time        string   `yaml:"time"`
	BQ          BQ       `yaml:"bq"`
	CS          CS       `yaml:"cs"`
	Reports     []Report `yaml:"reports"`
}

func NewScheduleConfig(filePath string) (*ScheduleConfig, error) {
	cfg := &ScheduleConfig{}

	err := cleanenv.ReadConfig(filePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
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

type CS struct {
	BucketName string `yaml:"bucket_name"`
}
