package rest

import (
	"MystiSql/internal/service/auth/oidc"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OIDCHandlers struct {
	flow      *oidc.OIDCFlow
	providers map[string]string
	logger    *zap.Logger
}

func NewOIDCHandlers(flow *oidc.OIDCFlow, providers map[string]string, logger *zap.Logger) *OIDCHandlers {
	if flow == nil {
		return nil
	}
	return &OIDCHandlers{
		flow:      flow,
		providers: providers,
		logger:    logger,
	}
}

func (h *OIDCHandlers) Login(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		if len(h.providers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no OIDC providers configured"})
			return
		}
		providerNames := make([]string, 0, len(h.providers))
		for name := range h.providers {
			providerNames = append(providerNames, name)
		}
		c.JSON(http.StatusOK, gin.H{"providers": providerNames})
		return
	}
	if _, exists := h.providers[providerName]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown OIDC provider: " + providerName})
		return
	}
	h.flow.Login(providerName, c.Writer, c.Request)
}

func (h *OIDCHandlers) Callback(c *gin.Context) {
	h.flow.Callback(c.Writer, c.Request)
}
