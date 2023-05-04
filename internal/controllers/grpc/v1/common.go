package v1

import (
	"time"
)

const msgErrMethod = "ошибка выполнения"
const msgMethodPrepared = "подготовка"
const msgMethodStarted = "запущено"
const msgMethodFinished = "завершено"

func pbDateNormalize(s string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
