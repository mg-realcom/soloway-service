package repository

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	solowaysdk "github.com/zfullio/soloway-sdk"
	"google.golang.org/api/sheets/v4"
	"soloway/internal/entity"
)

type IRepository interface {
	SendFromStorage(ctx context.Context, destination Destination, dateStart, dateFinish time.Time, bucketName string, file string, clientName string) (err error)
	GetUsers(ctx context.Context, spreadsheetID string) ([]entity.User, error)
	GetStatPlacementByDay(ctx context.Context, client *solowaysdk.Client, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error)
}

type IDB interface {
	ImportFromCS(ctx context.Context, destination Destination, bucket string, object string, schema any) (err error)
	CreateTable(ctx context.Context, destination Destination, schema any) (err error)
	TableExists(ctx context.Context, destination Destination) (err error)
	TruncateTable(ctx context.Context, destination Destination) (err error)
	DeleteByDateColumn(ctx context.Context, destination Destination, client string, dateColumn string, dateStart, dateFinish time.Time) error
}

type IStorage interface {
	SendFile(ctx context.Context, filename string) (err error)
}

type Repository struct {
	logger         *zerolog.Logger
	bd             IDB
	storage        IStorage
	spreadsheetSrv *sheets.Service
}

type Destination struct {
	ProjectID string
	DatasetID string
	TableID   string
}

func NewRepository(logger *zerolog.Logger, db IDB, storage IStorage, spreadsheetSrv *sheets.Service) (repository IRepository) {
	return Repository{
		logger:         logger,
		bd:             db,
		storage:        storage,
		spreadsheetSrv: spreadsheetSrv,
	}
}

type File struct {
	Name string
	Path string
	Date time.Time
}
