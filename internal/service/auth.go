package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
	"user-balance-api/internal/entity"
)

const (
	salt    = "poqi23ytoQUFOiwf82quof2qfoYFQUW"
	signKey = "OR*#Qiofhwfu3qru(*Q#rh2ir12q4QJWFO2"

	TokenTTL = 24 * time.Hour
)

type TokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo AuthRepo
}

func NewAuthService(repo AuthRepo) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(ctx context.Context, user entity.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)
	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) GenerateToken(ctx context.Context, username, password string) (string, error) {
	// get user from DB
	user, err := s.repo.GetUser(ctx, username, generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	// generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId: user.Id,
	})

	// sign token
	tokenString, err := token.SignedString([]byte(signKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(signKey), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return 0, fmt.Errorf("token claims are not of type TokenClaims")
	}

	return claims.UserId, nil
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
