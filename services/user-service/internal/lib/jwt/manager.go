package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type Manager struct {
	secret         []byte
	accessTokenTTL time.Duration
	issuer         string
}

type Claims struct {
	UserID int64  `json:"id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func New(cfg *config.Config) *Manager {
	return &Manager{
		secret:         []byte(cfg.JWT.Secret),
		accessTokenTTL: cfg.JWT.AccessTokenTTL,
		issuer:         cfg.JWT.Issuer,
	}
}

func (m *Manager) GenerateAccessToken(userID int64, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			Subject:   "access_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}

func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.Subject != "access_token" {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}
