package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTAuth handles JWT creation and validation.
type JWTAuth struct {
	secretKey []byte
}

// Claims defines the structure of the JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTAuth(secret string) *JWTAuth {
	return &JWTAuth{secretKey: []byte(secret)}
}

// GenerateToken creates a new JWT for a given user ID.
func (j *JWTAuth) GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(24 * 365 * time.Hour) // Long-lived for example
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken checks if a token string is valid.
func (j *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
