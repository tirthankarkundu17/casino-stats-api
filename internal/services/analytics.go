package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"tirthankarkundu17/casino-analytics/internal/models"
	"tirthankarkundu17/casino-analytics/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalyticsService struct {
	repo repository.TransactionRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repo repository.TransactionRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// GetGGR calculates the gross gaming revenue
func (s *AnalyticsService) GetGGR(ctx context.Context, from, to time.Time) ([]models.GGRResult, error) {
	return s.repo.GetGGR(ctx, from, to)
}

// GetDailyWagerVolume calculates the daily wager volume
func (s *AnalyticsService) GetDailyWagerVolume(ctx context.Context, from, to time.Time) ([]models.DailyVolumeResult, error) {
	results, err := s.repo.GetDailyWagerVolume(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Post-processing: set display fields from composite ID
	for i := range results {
		results[i].Date = results[i].ID.Date.Format("2006-01-02")
		results[i].Currency = results[i].ID.Currency
	}

	return results, nil
}

// GetUserWagerPercentileInMemory calculates the wager percentile of a user in memory
func (s *AnalyticsService) GetUserWagerPercentileInMemory(ctx context.Context, userID primitive.ObjectID, from, to time.Time) (*models.WagerPercentileResult, error) {
	allUsers, err := s.repo.GetAllUsersWagerRank(ctx, from, to)
	if err != nil {
		return nil, err
	}

	if len(allUsers) == 0 {
		return nil, &models.ErrNoData{Message: "no wagering data found in this timeframe"}
	}

	rank := -1
	var userTotal float64
	for i, u := range allUsers {
		if u.UserID == userID {
			rank = i + 1
			userTotal, _ = strconv.ParseFloat(u.TotalUSD.String(), 64)
			break
		}
	}

	if rank == -1 {
		return nil, &models.ErrNotFound{Resource: "user", Message: "user not found or has no wagers in this timeframe"}
	}

	percentile := (float64(rank) / float64(len(allUsers))) * 100

	return &models.WagerPercentileResult{
		UserID:          userID.Hex(),
		TotalUSDWagered: userTotal,
		Percentile:      percentile,
	}, nil
}

// GetUserWagerPercentile calculates the wager percentile of a user using database aggregation
func (s *AnalyticsService) GetUserWagerPercentile(ctx context.Context, userID primitive.ObjectID, from, to time.Time) (*models.WagerPercentileResult, error) {
	stats, err := s.repo.GetUserWagerStats(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	if stats == nil {
		return nil, &models.ErrNotFound{Resource: "user", Message: "user not found or has no wagers in this timeframe"}
	}

	percentile := (float64(stats.Rank) / float64(stats.TotalUsers)) * 100
	userTotal, err := strconv.ParseFloat(stats.TotalUSD.String(), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user total USD: %v", err)
	}

	return &models.WagerPercentileResult{
		UserID:          userID.Hex(),
		TotalUSDWagered: userTotal,
		Percentile:      percentile,
	}, nil
}
