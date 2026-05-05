package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	TenantID  string `json:"tenant_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func newTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Service encapsulates JWT signing and verification.
type Service struct {
	SigningKey []byte
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewService(signingKey []byte) (*Service, error) {
	if len(signingKey) < 32 {
		return nil, fmt.Errorf("signing key must be at least 32 bytes")
	}
	return &Service{
		SigningKey: signingKey,
		Issuer:     "goauth",
		AccessTTL:  1 * time.Hour,
		RefreshTTL: 30 * 24 * time.Hour,
	}, nil
}

// CreateAccessToken creates a signed JWT access token.
func (s *Service) CreateAccessToken(userID, tenantID string) (string, error) {
	return s.createJWT(userID, tenantID, "access", s.AccessTTL)
}

// CreateRefreshToken creates a signed JWT refresh token.
func (s *Service) CreateRefreshToken(userID, tenantID string) (string, error) {
	return s.createJWT(userID, tenantID, "refresh", s.RefreshTTL)
}

func (s *Service) createJWT(userID, tenantID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := newTokenID()
	if err != nil {
		return "", err
	}

	claims := &Claims{
		TenantID:  tenantID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.SigningKey)
}

func (s *Service) ParseAndValidate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
