package stats

type UserStats struct {
	UserID      int     `json:"userID"`
	Username    string  `json:"username"`
	GamesPlayed int     `json:"gamesPlayed"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	WinRate     float64 `json:"winRate"`
	TotalBoxes  int     `json:"totalBoxes"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   int    `json:"userID"`
	Username string `json:"username"`
	Value    int    `json:"value"`
}

type Leaderboard struct {
	MostWins  []LeaderboardEntry `json:"mostWins"`
	MostBoxes []LeaderboardEntry `json:"mostBoxes"`
}
