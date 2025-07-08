package user


type User struct {
	UserID int `json:"userID,omitempty"`
	Username string `json:"username"`
	Password string `json:"-"`
}


