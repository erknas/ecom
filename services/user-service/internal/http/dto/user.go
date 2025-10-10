package dto

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

type User struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type CreateUserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
}

type UpdateUserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (req CreateUserRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if errMsg := validateFirstName(req.FirstName); errMsg != "" {
		errors["first_name"] = errMsg
	}

	if !validateEmail(req.Email) {
		errors["email"] = "invalid email format"
	}

	if errMsg := validatePassword(req.Password); errMsg != "" {
		errors["password"] = errMsg
	}

	return errors
}

func (req UpdateUserRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if req.FirstName == nil && req.Email == nil && req.Password == nil {
		errors["empty_fields"] = "nothing to update"
		return errors
	}

	if req.FirstName != nil {
		if errMsg := validateFirstName(*req.FirstName); errMsg != "" {
			errors["first_name"] = errMsg
		}
	}

	if req.Email != nil {
		if !validateEmail(*req.Email) {
			errors["email"] = "invalid email format"
		}
	}

	if req.Password != nil {
		if errMsg := validatePassword(*req.Password); errMsg != "" {
			errors["password"] = errMsg
		}
	}

	return errors
}

func validateFirstName(name string) string {
	nameValue := strings.TrimSpace(name)

	if utf8.RuneCountInString(nameValue) == 0 {
		return "first name must not be empty"
	}

	if utf8.RuneCountInString(nameValue) < 2 {
		return "first name must be at least 2 characters"
	}

	if utf8.RuneCountInString(nameValue) > 16 {
		return "first name too long"
	}

	return ""
}

func validateEmail(email string) bool {
	emailValue := strings.TrimSpace(email)

	if len(emailValue) == 0 {
		return false
	}

	addr, err := mail.ParseAddress(emailValue)
	if err != nil {
		return false
	}

	parts := strings.Split(addr.Address, "@")

	local := parts[0]
	domain := parts[1]

	if len(local) == 0 {
		return false
	}

	if len(domain) == 0 {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	return true
}

func validatePassword(password string) string {
	if len(password) < 8 {
		return "password must be at least 8 characters long"
	}

	return ""
}
