package rest

import (
	"fmt"
	"net/http"

	"MystiSql/internal/service/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ValidatorHandlers struct {
	validatorService *validator.ValidatorService
	logger           *zap.Logger
}

func NewValidatorHandlers(validatorService *validator.ValidatorService, logger *zap.Logger) *ValidatorHandlers {
	return &ValidatorHandlers{
		validatorService: validatorService,
		logger:           logger,
	}
}

func (h *ValidatorHandlers) UpdateWhitelist(c *gin.Context) {
	var req struct {
		Patterns []string `json:"patterns" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	if err := h.validatorService.UpdateWhitelist(req.Patterns); err != nil {
		h.logger.Error("Failed to update whitelist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"UPDATE_FAILED",
			fmt.Sprintf("Failed to update whitelist: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Whitelist updated successfully",
		"count":   len(req.Patterns),
	})
}

func (h *ValidatorHandlers) GetWhitelist(c *gin.Context) {
	patterns := h.validatorService.GetWhitelist()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"patterns": patterns,
		"count":    len(patterns),
	})
}

func (h *ValidatorHandlers) UpdateBlacklist(c *gin.Context) {
	var req struct {
		Patterns []string `json:"patterns" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	if err := h.validatorService.UpdateBlacklist(req.Patterns); err != nil {
		h.logger.Error("Failed to update blacklist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"UPDATE_FAILED",
			fmt.Sprintf("Failed to update blacklist: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blacklist updated successfully",
		"count":   len(req.Patterns),
	})
}

func (h *ValidatorHandlers) GetBlacklist(c *gin.Context) {
	patterns := h.validatorService.GetBlacklist()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"patterns": patterns,
		"count":    len(patterns),
	})
}

func (h *ValidatorHandlers) ValidateQuery(c *gin.Context) {
	var req struct {
		Instance string `json:"instance" binding:"required"`
		Query    string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	result, err := h.validatorService.Validate(c.Request.Context(), req.Instance, req.Query)
	if err != nil {
		h.logger.Error("Failed to validate query", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"VALIDATION_FAILED",
			fmt.Sprintf("Failed to validate query: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"allowed":   result.Allowed,
		"reason":    result.Reason,
		"riskLevel": result.RiskLevel,
	})
}
