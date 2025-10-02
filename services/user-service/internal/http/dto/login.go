package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID          int64  `json:"id"`
	AccessToken string `json:"access_token"`
}
