package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenRevoked  = errors.New("token revoked")
	ErrInvalidClaims = errors.New("invalid claims")
)

type TokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenGenerator struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
	audience      string
}

func NewTokenGenerator(secretKey string, tokenDuration time.Duration) (*TokenGenerator, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	if tokenDuration <= 0 {
		return nil, errors.New("token duration must be positive")
	}

	return &TokenGenerator{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}, nil
}

// SetIssuerAudience 设置 JWT 的 issuer/audience claims（可选）。
// 设置后签发的 token 会携带这些 claims，且验证时会强制校验。
func (g *TokenGenerator) SetIssuerAudience(issuer, audience string) {
	g.issuer = issuer
	g.audience = audience
}

func (g *TokenGenerator) GenerateToken(userID, role string) (string, error) {
	if userID == "" {
		return "", errors.New("user ID cannot be empty")
	}
	if role == "" {
		return "", errors.New("role cannot be empty")
	}

	now := time.Now()
	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}
	if g.issuer != "" {
		claims.Issuer = g.issuer
	}
	if g.audience != "" {
		claims.Audience = []string{g.audience}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (g *TokenGenerator) ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	parserOptions := []jwt.ParserOption{}
	if g.issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(g.issuer))
	}
	if g.audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(g.audience))
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secretKey, nil
	}, parserOptions...)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}
