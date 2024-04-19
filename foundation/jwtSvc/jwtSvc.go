package jwtSvc

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
}

type JWTService interface {
	CreateToken(userID uuid.UUID, d time.Duration) (string, error)
	Verify(tokenS string) (*Claims, error)
}

type jwtSvc struct {
	key []byte
}

func NewJWTService(key []byte) (*jwtSvc, error) {
	if len(key) < 32 {
		return nil, errors.New("key minLength 32")
	}
	return &jwtSvc{key: key}, nil
}

func (j *jwtSvc) CreateToken(userID uuid.UUID, d time.Duration) (string, error) {
	claims := &Claims{}
	claims.Subject = userID.String()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(d))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenS, err := token.SignedString(j.key)
	if err != nil {
		return "", err
	}

	return tokenS, nil
}

func (j *jwtSvc) Verify(tokenS string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenS, &Claims{}, j.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("verify: invalid claims")
	}

	return claims, nil
}

func (j *jwtSvc) keyFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return []byte(j.key), nil
}

// func (j *JWT) checkSubject(userID string, claims Claims) error {
// 	if userID != claims["sub"] {
// 		return errors.New("invalid subject")
// 	}
// 	return nil
// }

// func (j *JWT) checkIssuer(claims Claims) error {
// 	issuer := os.Getenv("JWT_NOTES_ISSUER")
// 	if issuer != claims["iss"] {
// 		return errors.New("incorrect issuer")
// 	}
// 	return nil
// }

// func (j *JWT) checkExpSet(claims Claims) error {
// 	if _, ok := claims["exp"]; !ok {
// 		return fmt.Errorf("no expiration date set")
// 	}
// 	return nil
// }
