package rest

import (
	"context"
	"net/http"
	"time"

	"MystiSql/internal/service/batch"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BatchHandlers struct {
	batchService *batch.BatchService
	logger       *zap.Logger
}

func NewBatchHandlers(batchService *batch.BatchService, logger *zap.Logger) *BatchHandlers {
	return &BatchHandlers{
		batchService: batchService,
		logger:       logger,
	}
}

type BatchExecuteRequest struct {
	Instance       string   `json:"instance" binding:"required"`
	Queries        []string `json:"queries" binding:"required"`
	TransactionID  string   `json:"transactionId,omitempty"`
	StopOnError    bool     `json:"stopOnError"`
	UseTransaction bool     `json:"useTransaction"`
}

func (h *BatchHandlers) ExecuteBatch(c *gin.Context) {
	var req BatchExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	if len(req.Queries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Batch queries cannot be empty",
		})
		return
	}

	if len(req.Queries) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Batch size exceeds maximum limit of 1000 queries",
		})
		return
	}

	if req.Instance == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	batchReq := &batch.BatchRequest{
		Instance:      req.Instance,
		Queries:       req.Queries,
		TransactionID: req.TransactionID,
		StopOnError:   req.StopOnError,
	}

	var response *batch.BatchResponse
	var err error

	if req.UseTransaction && req.TransactionID == "" {
		response, err = h.batchService.ExecuteBatchWithNewTransaction(ctx, batchReq)
	} else {
		response, err = h.batchService.ExecuteBatch(ctx, batchReq)
	}

	if err != nil {
		h.logger.Error("Batch execution failed",
			zap.String("instance", req.Instance),
			zap.Int("query_count", len(req.Queries)),
			zap.Error(err),
		)

		if response != nil {
			c.JSON(http.StatusPartialContent, gin.H{
				"message": "Batch execution completed with errors",
				"result":  response,
				"error":   err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Batch execution failed",
				"details": err.Error(),
			})
		}
		return
	}

	h.logger.Info("Batch execution succeeded",
		zap.String("instance", req.Instance),
		zap.Int("success_count", response.SuccessCount),
		zap.Int("failure_count", response.FailureCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch execution completed successfully",
		"result":  response,
	})
}
