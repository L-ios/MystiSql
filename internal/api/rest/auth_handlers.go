package rest

import (
	"fmt"
	"net/http"

	"MystiSql/internal/service/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandlers struct {
	authService *auth.AuthService
	logger      *zap.Logger
}

func NewAuthHandlers(authService *auth.AuthService, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandlers) GenerateToken(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	token, err := h.authService.GenerateToken(c.Request.Context(), req.UserID, req.Role)
	if err != nil {
		h.logger.Error("Failed to generate token",
			zap.String("userId", req.UserID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"TOKEN_GENERATION_FAILED",
			fmt.Sprintf("Failed to generate token: %v", err),
		))
		return
	}

	tokenInfo, err := h.authService.GetTokenInfo(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("Failed to get token info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"TOKEN_INFO_FAILED",
			fmt.Sprintf("Failed to get token info: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, &GenerateTokenResponse{
		Success: true,
		Data: &TokenData{
			Token:     token,
			TokenID:   tokenInfo.TokenID,
			ExpiresAt: tokenInfo.ExpiresAt,
			IssuedAt:  tokenInfo.IssuedAt,
			UserID:    tokenInfo.UserID,
			Role:      tokenInfo.Role,
		},
	})
}

func (h *AuthHandlers) RevokeToken(c *gin.Context) {
	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	err := h.authService.RevokeToken(c.Request.Context(), req.Token)
	if err != nil {
		h.logger.Error("Failed to revoke token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"TOKEN_REVOCATION_FAILED",
			fmt.Sprintf("Failed to revoke token: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, &RevokeTokenResponse{
		Success: true,
		Message: "Token revoked successfully",
	})
}

func (h *AuthHandlers) ListTokens(c *gin.Context) {
	revokedTokens := h.authService.ListRevokedTokens(c.Request.Context())

	tokenInfos := make([]RevokedTokenInfo, 0, len(revokedTokens))
	for _, item := range revokedTokens {
		maskedToken := maskToken(item.Token)
		tokenInfos = append(tokenInfos, RevokedTokenInfo{
			Token:     maskedToken,
			Reason:    item.Reason,
			RevokedAt: item.RevokedAt,
		})
	}

	c.JSON(http.StatusOK, &TokensListResponse{
		Success:       true,
		RevokedTokens: tokenInfos,
	})
}

func (h *AuthHandlers) GetTokenInfo(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"MISSING_TOKEN",
			"Token parameter is required",
		))
		return
	}

	tokenInfo, err := h.authService.GetTokenInfo(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("Failed to get token info", zap.Error(err))

		statusCode := http.StatusUnauthorized
		errorCode := "INVALID_TOKEN"
		if err == auth.ErrTokenExpired {
			errorCode = "TOKEN_EXPIRED"
		} else if err == auth.ErrTokenRevoked {
			errorCode = "TOKEN_REVOKED"
		}

		c.JSON(statusCode, NewErrorResponse(
			errorCode,
			fmt.Sprintf("Failed to get token info: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, &TokenInfoResponse{
		Success:   true,
		UserID:    tokenInfo.UserID,
		Role:      tokenInfo.Role,
		TokenID:   tokenInfo.TokenID,
		ExpiresAt: tokenInfo.ExpiresAt,
		IssuedAt:  tokenInfo.IssuedAt,
	})
}

func maskToken(token string) string {
	if len(token) <= 20 {
		return "****"
	}
	return token[:10] + "..." + token[len(token)-10:]
}
