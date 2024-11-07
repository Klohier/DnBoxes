package user


type User struct {
	UserID int `json:"userID,omitempty"`
	Email string  `json:"email,omitempty"`
	Username string `json:"username"`
	Password string `json:"-"`
	IsInGame bool `json:"isInGsme"`
	CurrentGameID string `json:"currentGameID,omitempty"`

}


