package sol

import (
	"Soloway/internal/domain/entity"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zfullio/soloway-sdk"
)

type PlacementRepository struct {
	client solowaysdk.Client
	logger *zerolog.Logger
}

func NewPlacementRepository(client solowaysdk.Client, logger *zerolog.Logger) *PlacementRepository {
	repoLogger := logger.With().Str("repository", "placement").Str("storage", "soloway").Logger()

	return &PlacementRepository{client: client, logger: &repoLogger}
}

func (ps PlacementRepository) GetAll(ctx context.Context) ([]entity.Placement, error) {
	ps.logger.Trace().Msg("GetAll")

	data, err := ps.client.GetPlacements(ctx)

	if err != nil {
		return nil, fmt.Errorf("PlacementRepository.GetAll: %w", err)
	}

	placements := make([]entity.Placement, 0, len(data.List))

	for i := 0; i < len(data.List); i++ {
		placements = append(placements, *newPlacement(data.List[i]))
	}

	return placements, nil
}

func newPlacement(placement solowaysdk.Placement) *entity.Placement {
	return &entity.Placement{
		GUID: placement.Doc.GUID,
		Name: placement.Doc.Name,
	}
}
