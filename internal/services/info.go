package services

import (
	"solowayStat/internal/repository"
	"solowayStat/pkg/solowaysdk"
)

type InfoService struct {
	client solowaysdk.Client
	repo   repository.Repository
}

func NewInfoService(client solowaysdk.Client, repo repository.Repository) *InfoService {
	return &InfoService{
		client: client,
		repo:   repo,
	}
}

func (s *InfoService) Login() (err error) {
	err = s.client.Login()
	if err != nil {
		return err
	}
	return nil
}
func (s *InfoService) Whoami() (err error) {
	err = s.client.Whoami()
	if err != nil {
		return err
	}
	return nil
}
func (s *InfoService) GetPlacements() (placements solowaysdk.PlacementsInfo, err error) {
	placements, err = s.client.GetPlacements()
	if err != nil {
		return solowaysdk.PlacementsInfo{}, err
	}
	return placements, nil
}
