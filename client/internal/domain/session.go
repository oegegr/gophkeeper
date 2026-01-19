package domain

type Session struct {
    UserID   string    `json:"user_id"`
    Username string    `json:"username"`
    Token    string    `json:"token"`
}