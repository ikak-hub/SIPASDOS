package helper

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sipasdos/app/model"
)

const (
	RoleAdmin    = model.RoleAdmin
	RoleLecturer = model.RoleLecturer
	RoleStudent  = model.RoleStudent
)

const (
	LocalUserID = "auth_user_id"
	LocalRole   = "auth_role"
)

const tokenIssuer = "sipasdos"

var ErrInvalidToken = errors.New("token tidak valid atau kedaluwarsa")

type Claims struct {
	UserID int    `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func GenerateToken(secret string, ttl time.Duration, userID int, role string) (string, time.Time, error) {
	now := time.Now()
	expires := now.Add(ttl)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

func ParseToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func CurrentUser(c *fiber.Ctx) (int, string, bool) {
	id, okID := c.Locals(LocalUserID).(int)
	role, okRole := c.Locals(LocalRole).(string)
	return id, role, okID && okRole
}
