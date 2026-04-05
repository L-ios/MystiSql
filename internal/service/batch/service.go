package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection/pool"
	"MystiSql/internal/service/transaction"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type BatchResult struct {
	Index         int    `json:"index"`
	SQL           string `json:"sql"`
	RowsAffected  int64  `json:"rowsAffected"`
	LastInsertID  int64  `json:"lastInsertId,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	ExecutionTime int64  `json:"executionTimeMs"`
}

type BatchResponse struct {
	Results            []BatchResult `json:"results"`
	TotalRowsAffected  int64         `json:"totalRowsAffected"`
	SuccessCount       int           `json:"successCount"`
	FailureCount       int           `json:"failureCount"`
	TotalExecutionTime int64         `json:"totalExecutionTimeMs"`
}

type BatchConfig struct {
	MaxBatchSize   int           `yaml:"maxBatchSize"`
	EnableParallel bool          `yaml:"enableParallel"`
	MaxWorkers     int           `yaml:"maxWorkers"`
	Timeout        time.Duration `yaml:"timeout"`
}

func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		MaxBatchSize:   1000,
		EnableParallel: false,
		MaxWorkers:     10,
		Timeout:        5 * time.Minute,
	}
}

type BatchService struct {
	txManager   *transaction.TransactionManager
	poolManager *pool.ConnectionPoolManager
	config      *BatchConfig
	logger      *zap.Logger
}

func NewBatchService(txManager *transaction.TransactionManager, poolManager *pool.ConnectionPoolManager, config *BatchConfig, logger *zap.Logger) *BatchService {
	if config == nil {
		config = DefaultBatchConfig()
	}

	return &BatchService{
		txManager:   txManager,
		poolManager: poolManager,
		config:      config,
		logger:      logger,
	}
}

type BatchRequest struct {
	Instance      string   `json:"instance" binding:"required"`
	Queries       []string `json:"queries" binding:"required"`
	TransactionID string   `json:"transactionId,omitempty"`
	StopOnError   bool     `json:"stopOnError"`
}

func (s *BatchService) ExecuteBatch(ctx context.Context, req *BatchRequest) (*BatchResponse, error) {
	if len(req.Queries) == 0 {
		return nil, fmt.Errorf("batch queries cannot be empty")
	}

	if len(req.Queries) > s.config.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum limit %d", len(req.Queries), s.config.MaxBatchSize)
	}

	s.logger.Info("Starting batch execution",
		zap.String("instance", req.Instance),
		zap.Int("query_count", len(req.Queries)),
		zap.String("transaction_id", req.TransactionID),
	)

	start := time.Now()
	results := make([]BatchResult, len(req.Queries))
	var totalRowsAffected int64
	var successCount, failureCount int
	var mu sync.Mutex

	if s.config.EnableParallel && req.TransactionID == "" {
		s.executeBatchParallel(ctx, req, results, &mu, &totalRowsAffected, &successCount, &failureCount)
	} else {
		s.executeBatchSequential(ctx, req, results, &totalRowsAffected, &successCount, &failureCount)
	}

	totalTime := time.Since(start).Milliseconds()

	s.logger.Info("Batch execution completed",
		zap.Int("success", successCount),
		zap.Int("failure", failureCount),
		zap.Int64("total_time_ms", totalTime),
	)

	return &BatchResponse{
		Results:            results,
		TotalRowsAffected:  totalRowsAffected,
		SuccessCount:       successCount,
		FailureCount:       failureCount,
		TotalExecutionTime: totalTime,
	}, nil
}

func (s *BatchService) executeBatchSequential(
	ctx context.Context,
	req *BatchRequest,
	results []BatchResult,
	totalRowsAffected *int64,
	successCount, failureCount *int,
) {
	for i, sql := range req.Queries {
		result := s.executeSingleQuery(ctx, req, i, sql)
		results[i] = result

		if result.Success {
			*totalRowsAffected += result.RowsAffected
			*successCount++
		} else {
			*failureCount++
			if req.StopOnError {
				s.logger.Warn("Stopping batch execution on error",
					zap.Int("index", i),
					zap.String("error", result.Error),
				)
				for j := i + 1; j < len(req.Queries); j++ {
					results[j] = BatchResult{
						Index:   j,
						SQL:     req.Queries[j],
						Success: false,
						Error:   "skipped due to previous error",
					}
					*failureCount++
				}
				break
			}
		}
	}
}

func (s *BatchService) executeBatchParallel(
	ctx context.Context,
	req *BatchRequest,
	results []BatchResult,
	mu *sync.Mutex,
	totalRowsAffected *int64,
	successCount, failureCount *int,
) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.config.MaxWorkers)

	for i, sql := range req.Queries {
		wg.Add(1)
		go func(index int, query string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.executeSingleQuery(ctx, req, index, query)

			mu.Lock()
			results[index] = result
			if result.Success {
				*totalRowsAffected += result.RowsAffected
				*successCount++
			} else {
				*failureCount++
			}
			mu.Unlock()
		}(i, sql)
	}

	wg.Wait()
}

func (s *BatchService) executeSingleQuery(ctx context.Context, req *BatchRequest, index int, sql string) BatchResult {
	start := time.Now()

	result := BatchResult{
		Index: index,
		SQL:   sql,
	}

	if req.TransactionID != "" && s.txManager != nil {
		tx, err := s.txManager.GetTransaction(req.TransactionID)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to get transaction: %v", err)
			result.ExecutionTime = time.Since(start).Milliseconds()
			return result
		}

		execResult, err := tx.Connection.Exec(ctx, sql)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.RowsAffected = execResult.RowsAffected
			result.LastInsertID = execResult.LastInsertID
		}
	} else if s.poolManager != nil {
		conn, err := s.poolManager.GetConnection(ctx, req.Instance)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to get connection: %v", err)
		} else {
			defer conn.Close()

			execResult, execErr := conn.Exec(ctx, sql)
			if execErr != nil {
				result.Success = false
				result.Error = execErr.Error()
			} else {
				result.Success = true
				result.RowsAffected = execResult.RowsAffected
				result.LastInsertID = execResult.LastInsertID
			}
		}
	} else {
		result.Success = false
		result.Error = "no connection available: neither transaction manager nor pool manager is configured"
	}

	result.ExecutionTime = time.Since(start).Milliseconds()
	return result
}

func (s *BatchService) ExecuteBatchWithNewTransaction(ctx context.Context, req *BatchRequest) (*BatchResponse, error) {
	if s.txManager == nil {
		return nil, fmt.Errorf("transaction manager not available")
	}

	tx, err := s.txManager.BeginTransaction(ctx, req.Instance, types.IsolationLevelDefault, "batch-user")
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	txReq := &BatchRequest{
		Instance:      req.Instance,
		Queries:       req.Queries,
		TransactionID: tx.ID,
		StopOnError:   req.StopOnError,
	}

	response, err := s.ExecuteBatch(ctx, txReq)
	if err != nil {
		_ = s.txManager.RollbackTransaction(ctx, tx.ID)
		return nil, err
	}

	if response.FailureCount > 0 {
		_ = s.txManager.RollbackTransaction(ctx, tx.ID)
		return response, fmt.Errorf("batch execution failed with %d errors", response.FailureCount)
	}

	if err := s.txManager.CommitTransaction(ctx, tx.ID); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return response, nil
}
