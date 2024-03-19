package converters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/civil"
	"soloway/internal/entity"
)

func GeneratePlacementStatByDayJSON(attachmentDir string, stat []entity.StatPlacement, clientName string, dateUpdate time.Time) (string, error) {
	t := time.Now()
	fileName := fmt.Sprintf("%s_%s", "placement_stat", t.Format("2006-01-02_15:04:05"))

	file, err := os.CreateTemp(attachmentDir, fmt.Sprintf("%s.*.json", fileName))
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error close file: %s", err)
		}
	}(file)

	statJSON := make([]StatPlacementJSON, 0, len(stat))

	for _, item := range stat {
		conv, err := ConvertPlacementToJSON(item, clientName, dateUpdate)
		if err != nil {
			return "", fmt.Errorf("error convert: %w", err)
		}

		statJSON = append(statJSON, conv)
	}

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	for _, item := range statJSON {
		err := encoder.Encode(item)
		if err != nil {
			return "", err
		}
	}

	err = writer.Flush()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

type StatPlacementJSON struct {
	ClientName    string     `json:"client_name"`
	Clicks        int64      `json:"clicks"`
	Cost          int64      `json:"cost"`
	PlacementID   string     `json:"placement_id"`
	PlacementName string     `json:"placement_name"`
	Exposures     int64      `json:"exposures"`
	Date          civil.Date `json:"date"`
	DateUpdate    string     `json:"date_update"`
}

func ConvertPlacementToJSON(placement entity.StatPlacement, clientName string, dateUpdate time.Time) (StatPlacementJSON, error) {
	return StatPlacementJSON{
		ClientName:    clientName,
		Clicks:        int64(placement.Clicks),
		Cost:          int64(placement.Cost),
		PlacementID:   placement.PlacementID,
		PlacementName: placement.PlacementName,
		Exposures:     int64(placement.Exposures),
		Date:          civil.DateOf(placement.Date),
		DateUpdate:    dateUpdate.Format(time.DateTime),
	}, nil
}
