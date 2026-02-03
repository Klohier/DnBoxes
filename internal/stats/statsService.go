package stats

import "context"

type StatsService struct {
	repo *PgStatsRepository
}

func NewStatsService(repo *PgStatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetUserStats(ctx context.Context, userID int) (*UserStats, error) {
	return s.repo.GetUserStats(ctx, userID)
}

func (s *StatsService) GetLeaderboard(ctx context.Context, limit int) (*Leaderboard, error) {
	return s.repo.GetLeaderboard(ctx, limit)
}
