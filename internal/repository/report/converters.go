package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"soloway/internal/entity"
)

func GenerateReportPlacementStatJSON(attachmentDir string, stat []entity.StatPlacement) ([]entity.File, error) {
	groupedData := make(map[string][]entity.StatPlacement)

	for _, item := range stat {
		groupedData[item.Date] = append(groupedData[item.Date], item)
	}

	var files []entity.File

	for date, rows := range groupedData {
		fileName := fmt.Sprintf("%s_%s", "calls", date)

		file, err := writeFile(attachmentDir, fileName, rows)
		if err != nil {
			return nil, err
		}

		dt, err := time.Parse(time.DateOnly, date)
		if err != nil {
			return nil, err
		}

		files = append(files, entity.File{
			Name: fileName,
			Path: file.Name(),
			Date: dt,
		})
	}

	return files, nil
}

func writeFile[T any](attachmentDir, fileName string, data []T) (*os.File, error) {
	file, err := os.CreateTemp(attachmentDir, fmt.Sprintf("%s.*.json", fileName))
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error close file: %s", err)
		}
	}(file)

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	for _, item := range data {
		err := encoder.Encode(item)
		if err != nil {
			_ = os.Remove(file.Name())

			return nil, err
		}
	}

	err = writer.Flush()
	if err != nil {
		_ = os.Remove(file.Name())

		return nil, err
	}

	return file, nil
}
