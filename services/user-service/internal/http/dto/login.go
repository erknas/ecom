package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID          int64  `json:"id"`
	AccessToken string `json:"access_token"`
}

func (req LoginRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if req.Email == "" {
		errors["email"] = "email required"
	}

	if req.Password == "" {
		errors["password"] = "password required"
	}

	return errors
}
