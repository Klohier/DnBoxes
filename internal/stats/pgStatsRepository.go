package stats

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStatsRepository struct {
	db *pgxpool.Pool
}

func NewPgStatsRepository(db *pgxpool.Pool) *PgStatsRepository {
	return &PgStatsRepository{db: db}
}

func (repo *PgStatsRepository) GetUserStats(ctx context.Context, userID int) (*UserStats, error) {
	var stats UserStats
	stats.UserID = userID

	// Get username
	err := repo.db.QueryRow(ctx, `SELECT username FROM users WHERE user_id = $1`, userID).Scan(&stats.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user %d: %w", userID, err)
	}

	// Get games played and total boxes from completed games
	err = repo.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT g.game_id), COALESCE(SUM(gd.score), 0)
		FROM game_details gd
		JOIN games g ON g.game_id = gd.game_id
		WHERE gd.user_id = $1 AND g.ended_at IS NOT NULL
	`, userID).Scan(&stats.GamesPlayed, &stats.TotalBoxes)
	if err != nil {
		return nil, fmt.Errorf("failed to get game stats for user %d: %w", userID, err)
	}

	// Get wins
	err = repo.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM games
		WHERE winner_id = $1 AND ended_at IS NOT NULL
	`, userID).Scan(&stats.Wins)
	if err != nil {
		return nil, fmt.Errorf("failed to get win count for user %d: %w", userID, err)
	}

	stats.Losses = stats.GamesPlayed - stats.Wins
	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.GamesPlayed) * 100
	}

	return &stats, nil
}

func (repo *PgStatsRepository) GetLeaderboard(ctx context.Context, limit int) (*Leaderboard, error) {
	if limit <= 0 {
		limit = 10
	}

	leaderboard := &Leaderboard{}

	// Most wins
	rows, err := repo.db.Query(ctx, `
		SELECT u.user_id, u.username, COUNT(g.game_id) as wins
		FROM users u
		JOIN games g ON g.winner_id = u.user_id AND g.ended_at IS NOT NULL
		WHERE u.is_guest = false
		GROUP BY u.user_id, u.username
		ORDER BY wins DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query wins leaderboard: %w", err)
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Value); err != nil {
			return nil, fmt.Errorf("failed to scan wins entry: %w", err)
		}
		entry.Rank = rank
		rank++
		leaderboard.MostWins = append(leaderboard.MostWins, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Most boxes completed
	rows2, err := repo.db.Query(ctx, `
		SELECT u.user_id, u.username, COALESCE(SUM(gd.score), 0)::int as total_boxes
		FROM users u
		JOIN game_details gd ON gd.user_id = u.user_id
		JOIN games g ON g.game_id = gd.game_id AND g.ended_at IS NOT NULL
		WHERE u.is_guest = false
		GROUP BY u.user_id, u.username
		ORDER BY total_boxes DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query boxes leaderboard: %w", err)
	}
	defer rows2.Close()

	rank = 1
	for rows2.Next() {
		var entry LeaderboardEntry
		if err := rows2.Scan(&entry.UserID, &entry.Username, &entry.Value); err != nil {
			return nil, fmt.Errorf("failed to scan boxes entry: %w", err)
		}
		entry.Rank = rank
		rank++
		leaderboard.MostBoxes = append(leaderboard.MostBoxes, entry)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return leaderboard, nil
}
