package service

import (
	"context"

	"github.com/creafly/identity/internal/domain/repository"
)

type AnalyticsResult struct {
	Users *repository.UserAnalytics `json:"users"`
}

type AnalyticsService interface {
	GetAnalytics(ctx context.Context) (*AnalyticsResult, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

func (s *analyticsService) GetAnalytics(ctx context.Context) (*AnalyticsResult, error) {
	userAnalytics, err := s.analyticsRepo.GetUserAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	return &AnalyticsResult{
		Users: userAnalytics,
	}, nil
}
