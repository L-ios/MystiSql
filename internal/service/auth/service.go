package auth

import (
	"context"
	"fmt"
	"time"
)

type AuthService struct {
	generator *TokenGenerator
	blacklist *TokenBlacklist
}

type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	TokenID   string    `json:"token_id"`
}

func NewAuthService(secretKey string, tokenDuration time.Duration) (*AuthService, error) {
	generator, err := NewTokenGenerator(secretKey, tokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create token generator: %w", err)
	}

	return &AuthService{
		generator: generator,
		blacklist: NewTokenBlacklist(),
	}, nil
}

func (s *AuthService) GenerateToken(ctx context.Context, userID, role string) (string, error) {
	token, err := s.generator.GenerateToken(userID, role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if s.blacklist.Contains(tokenString) {
		return nil, ErrTokenRevoked
	}

	claims, err := s.generator.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, tokenString string) error {
	claims, err := s.generator.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	s.blacklist.Add(tokenString, fmt.Sprintf("revoked by user %s", claims.UserID))
	return nil
}

func (s *AuthService) IsTokenRevoked(ctx context.Context, tokenString string) bool {
	return s.blacklist.Contains(tokenString)
}

func (s *AuthService) GetTokenInfo(ctx context.Context, tokenString string) (*TokenInfo, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		UserID:    claims.UserID,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
		TokenID:   claims.ID,
	}, nil
}

func (s *AuthService) ListRevokedTokens(ctx context.Context) []*BlacklistItem {
	return s.blacklist.GetAll()
}
