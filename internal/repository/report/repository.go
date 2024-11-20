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
	UploadToStorage(ctx context.Context, directory string, bucketName string, filePath string, date time.Time) (err error)
	StorageClearByDate(ctx context.Context, directory string, bucketName string, date time.Time) (err error)
	GetUsers(ctx context.Context, spreadsheetID string) ([]entity.User, error)
	GetStatPlacementByDay(ctx context.Context, client *solowaysdk.Client, startDate time.Time, stopDate time.Time, attachmentDir string) ([]entity.File, error)
}

type IStorage interface {
	UploadFileWithDateDestination(ctx context.Context, bucketName string, directory string, filePath string, date time.Time) error
	DeleteFolderByDate(ctx context.Context, bucketName string, directory string, date time.Time) error
}

type Repository struct {
	logger         *zerolog.Logger
	storage        IStorage
	spreadsheetSrv *sheets.Service
}

type Destination struct {
	ProjectID string
	DatasetID string
	TableID   string
}

func NewRepository(logger *zerolog.Logger, storage IStorage, spreadsheetSrv *sheets.Service) (repository IRepository) {
	return Repository{
		logger:         logger,
		storage:        storage,
		spreadsheetSrv: spreadsheetSrv,
	}
}
