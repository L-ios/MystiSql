package rest

import (
	"net/http"
	"time"

	"MystiSql/internal/service/transaction"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TransactionHandlers struct {
	txManager *transaction.TransactionManager
	logger    *zap.Logger
}

func NewTransactionHandlers(txManager *transaction.TransactionManager, logger *zap.Logger) *TransactionHandlers {
	return &TransactionHandlers{
		txManager: txManager,
		logger:    logger,
	}
}

type BeginTransactionRequest struct {
	Instance       string               `json:"instance" binding:"required"`
	IsolationLevel types.IsolationLevel `json:"isolation_level,omitempty"`
	Timeout        string               `json:"timeout,omitempty"`
}

type BeginTransactionResponse struct {
	TransactionID string    `json:"transaction_id"`
	ConnectionID  string    `json:"connection_id"`
	Instance      string    `json:"instance"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type CommitTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type RollbackTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type TransactionQueryRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Query         string `json:"query" binding:"required"`
}

type TransactionInfoResponse struct {
	TransactionID  string                       `json:"transaction_id"`
	ConnectionID   string                       `json:"connection_id"`
	Instance       string                       `json:"instance"`
	State          transaction.TransactionState `json:"state"`
	IsolationLevel types.IsolationLevel         `json:"isolation_level"`
	CreatedAt      time.Time                    `json:"created_at"`
	ExpiresAt      time.Time                    `json:"expires_at"`
	LastActivityAt time.Time                    `json:"last_activity_at"`
	UserID         string                       `json:"user_id"`
}

func (h *TransactionHandlers) BeginTransaction(c *gin.Context) {
	var req BeginTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "anonymous"
	}

	// Parse timeout (will be handled by transaction manager)
	if req.Timeout != "" {
		// Timeout is specified in request, will be used by manager
		h.logger.Debug("Custom timeout specified", zap.String("timeout", req.Timeout))
	}

	// Begin transaction
	tx, err := h.txManager.BeginTransaction(
		c.Request.Context(),
		req.Instance,
		req.IsolationLevel,
		userID.(string),
	)
	if err != nil {
		h.logger.Error("Failed to begin transaction",
			zap.String("instance", req.Instance),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to begin transaction",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Transaction started",
		zap.String("tx_id", tx.ID),
		zap.String("instance", req.Instance),
		zap.String("user_id", userID.(string)),
	)

	c.JSON(http.StatusOK, BeginTransactionResponse{
		TransactionID: tx.ID,
		ConnectionID:  tx.ConnectionID,
		Instance:      tx.Instance,
		CreatedAt:     tx.CreatedAt,
		ExpiresAt:     tx.ExpiresAt,
	})
}

func (h *TransactionHandlers) CommitTransaction(c *gin.Context) {
	var req CommitTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	err := h.txManager.CommitTransaction(c.Request.Context(), req.TransactionID)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		if err == transaction.ErrTransactionExpired {
			c.JSON(http.StatusGone, gin.H{"error": "Transaction expired"})
			return
		}
		if err == transaction.ErrTransactionNotActive {
			c.JSON(http.StatusConflict, gin.H{"error": "Transaction is not active"})
			return
		}

		h.logger.Error("Failed to commit transaction",
			zap.String("tx_id", req.TransactionID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Transaction committed",
		zap.String("tx_id", req.TransactionID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Transaction committed successfully",
		"transaction_id": req.TransactionID,
	})
}

func (h *TransactionHandlers) RollbackTransaction(c *gin.Context) {
	var req RollbackTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	err := h.txManager.RollbackTransaction(c.Request.Context(), req.TransactionID)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		if err == transaction.ErrTransactionNotActive {
			c.JSON(http.StatusConflict, gin.H{"error": "Transaction is not active"})
			return
		}

		h.logger.Error("Failed to rollback transaction",
			zap.String("tx_id", req.TransactionID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rollback transaction",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Transaction rolled back",
		zap.String("tx_id", req.TransactionID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Transaction rolled back successfully",
		"transaction_id": req.TransactionID,
	})
}

func (h *TransactionHandlers) GetTransaction(c *gin.Context) {
	txID := c.Param("id")

	tx, err := h.txManager.GetTransaction(txID)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		if err == transaction.ErrTransactionExpired {
			c.JSON(http.StatusGone, gin.H{"error": "Transaction expired"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TransactionInfoResponse{
		TransactionID:  tx.ID,
		ConnectionID:   tx.ConnectionID,
		Instance:       tx.Instance,
		State:          tx.State,
		IsolationLevel: tx.IsolationLevel,
		CreatedAt:      tx.CreatedAt,
		ExpiresAt:      tx.ExpiresAt,
		LastActivityAt: tx.LastActivityAt,
		UserID:         tx.UserID,
	})
}

func (h *TransactionHandlers) ListTransactions(c *gin.Context) {
	transactions := h.txManager.ListTransactions()

	response := make([]TransactionInfoResponse, 0, len(transactions))
	for _, tx := range transactions {
		response = append(response, TransactionInfoResponse{
			TransactionID:  tx.ID,
			ConnectionID:   tx.ConnectionID,
			Instance:       tx.Instance,
			State:          tx.State,
			IsolationLevel: tx.IsolationLevel,
			CreatedAt:      tx.CreatedAt,
			ExpiresAt:      tx.ExpiresAt,
			LastActivityAt: tx.LastActivityAt,
			UserID:         tx.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": response,
		"count":        len(response),
	})
}

func (h *TransactionHandlers) ExtendTransaction(c *gin.Context) {
	txID := c.Param("id")

	var req struct {
		Duration string `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid duration format",
			"details": err.Error(),
		})
		return
	}

	err = h.txManager.ExtendTransaction(txID, duration)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		if err == transaction.ErrTransactionNotActive {
			c.JSON(http.StatusConflict, gin.H{"error": "Transaction is not active"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to extend transaction",
			"details": err.Error(),
		})
		return
	}

	// Get updated transaction
	tx, _ := h.txManager.GetTransaction(txID)

	h.logger.Info("Transaction extended",
		zap.String("tx_id", txID),
		zap.Duration("duration", duration),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Transaction extended successfully",
		"transaction_id": txID,
		"new_expires_at": tx.ExpiresAt,
	})
}
