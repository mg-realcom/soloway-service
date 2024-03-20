package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type ScheduleConfig struct {
	Destination string   `yaml:"destination"`
	Time        string   `yaml:"time"`
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
	BucketName       string `yaml:"bucket_name"`
	Days             int    `yaml:"period"`
}
