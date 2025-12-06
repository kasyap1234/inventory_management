package common

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManager provides utilities for managing database transactions
type TransactionManager struct {
	pool   *pgxpool.Pool
	logger *StructuredLogger
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(pool *pgxpool.Pool, logger *StructuredLogger) *TransactionManager {
	return &TransactionManager{
		pool:   pool,
		logger: logger,
	}
}

// TransactionFunc is a function that executes within a transaction
type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

// ExecuteInTransaction executes a function within a database transaction
// It automatically handles commit/rollback based on the function's return value
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn TransactionFunc) error {
	// Begin transaction
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		tm.logger.ErrorWithContext(ctx, "Failed to begin transaction", err, nil)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Propagate transaction in context so repositories can reuse it
	ctx = context.WithValue(ctx, TransactionKey, tx)

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			tm.logger.ErrorWithContext(ctx, "Transaction panicked, rolled back", fmt.Errorf("panic: %v", p), nil)
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(ctx, tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			tm.logger.ErrorWithContext(ctx, "Failed to rollback transaction", rbErr, map[string]interface{}{
				"original_error": err.Error(),
			})
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		tm.logger.DebugWithContext(ctx, "Transaction rolled back due to error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		tm.logger.ErrorWithContext(ctx, "Failed to commit transaction", err, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tm.logger.DebugWithContext(ctx, "Transaction committed successfully", nil)
	return nil
}

// ExecuteInTransactionWithRetry executes a function within a transaction with retry logic
// It retries on serialization failures or deadlocks
func (tm *TransactionManager) ExecuteInTransactionWithRetry(ctx context.Context, maxRetries int, fn TransactionFunc) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			tm.logger.InfoWithContext(ctx, "Retrying transaction", map[string]interface{}{
				"attempt":     attempt,
				"max_retries": maxRetries,
			})
		}

		err := tm.ExecuteInTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable (serialization failure or deadlock)
		if !isRetryableError(err) {
			return err
		}

		if attempt == maxRetries {
			tm.logger.WarnWithContext(ctx, "Transaction failed after max retries", map[string]interface{}{
				"attempts": attempt + 1,
				"error":    err.Error(),
			})
			break
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError checks if an error is retryable (serialization failure or deadlock)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for PostgreSQL serialization errors
	errMsg := err.Error()
	return contains(errMsg, "serialization failure") ||
		contains(errMsg, "deadlock detected") ||
		contains(errMsg, "could not serialize access")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TransactionOptions provides options for transaction execution
type TransactionOptions struct {
	IsolationLevel pgx.TxIsoLevel
	AccessMode     pgx.TxAccessMode
	DeferrableMode pgx.TxDeferrableMode
}

// ExecuteInTransactionWithOptions executes a function within a transaction with custom options
func (tm *TransactionManager) ExecuteInTransactionWithOptions(ctx context.Context, opts TransactionOptions, fn TransactionFunc) error {
	// Begin transaction with options
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       opts.IsolationLevel,
		AccessMode:     opts.AccessMode,
		DeferrableMode: opts.DeferrableMode,
	})
	if err != nil {
		tm.logger.ErrorWithContext(ctx, "Failed to begin transaction with options", err, map[string]interface{}{
			"isolation_level": opts.IsolationLevel,
		})
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Propagate transaction in context so repositories can reuse it
	ctx = context.WithValue(ctx, TransactionKey, tx)

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			tm.logger.ErrorWithContext(ctx, "Transaction panicked, rolled back", fmt.Errorf("panic: %v", p), nil)
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			tm.logger.ErrorWithContext(ctx, "Failed to rollback transaction", rbErr, map[string]interface{}{
				"original_error": err.Error(),
			})
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		tm.logger.ErrorWithContext(ctx, "Failed to commit transaction", err, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Savepoint represents a transaction savepoint
type Savepoint struct {
	name string
	tx   pgx.Tx
}

// CreateSavepoint creates a savepoint within a transaction
func CreateSavepoint(ctx context.Context, tx pgx.Tx, name string) (*Savepoint, error) {
	_, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}

	return &Savepoint{
		name: name,
		tx:   tx,
	}, nil
}

// Rollback rolls back to this savepoint
func (s *Savepoint) Rollback(ctx context.Context) error {
	_, err := s.tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", s.name))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", s.name, err)
	}
	return nil
}

// Release releases this savepoint
func (s *Savepoint) Release(ctx context.Context) error {
	_, err := s.tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", s.name))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", s.name, err)
	}
	return nil
}

// BatchOperation represents a batch of operations to execute in a transaction
type BatchOperation struct {
	Name      string
	Operation func(ctx context.Context, tx pgx.Tx) error
	OnError   func(ctx context.Context, err error) error // Optional error handler
}

// ExecuteBatch executes multiple operations in a single transaction
// If any operation fails, all operations are rolled back
func (tm *TransactionManager) ExecuteBatch(ctx context.Context, operations []BatchOperation) error {
	return tm.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		for i, op := range operations {
			tm.logger.DebugWithContext(ctx, "Executing batch operation", map[string]interface{}{
				"operation": op.Name,
				"index":     i,
			})

			if err := op.Operation(ctx, tx); err != nil {
				tm.logger.ErrorWithContext(ctx, "Batch operation failed", err, map[string]interface{}{
					"operation": op.Name,
					"index":     i,
				})

				// Call error handler if provided
				if op.OnError != nil {
					if handlerErr := op.OnError(ctx, err); handlerErr != nil {
						return fmt.Errorf("operation %s failed: %w, error handler failed: %v", op.Name, err, handlerErr)
					}
				}

				return fmt.Errorf("operation %s failed: %w", op.Name, err)
			}
		}

		tm.logger.InfoWithContext(ctx, "All batch operations completed successfully", map[string]interface{}{
			"total_operations": len(operations),
		})

		return nil
	})
}

// ExecuteBatchWithSavepoints executes multiple operations with savepoints
// If an operation fails, it can be rolled back to the savepoint without affecting previous operations
func (tm *TransactionManager) ExecuteBatchWithSavepoints(ctx context.Context, operations []BatchOperation, continueOnError bool) error {
	return tm.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var errors []error

		for i, op := range operations {
			// Create savepoint before each operation
			savepoint, err := CreateSavepoint(ctx, tx, fmt.Sprintf("sp_%d", i))
			if err != nil {
				return fmt.Errorf("failed to create savepoint for operation %s: %w", op.Name, err)
			}

			tm.logger.DebugWithContext(ctx, "Executing batch operation with savepoint", map[string]interface{}{
				"operation": op.Name,
				"index":     i,
			})

			// Execute operation
			if err := op.Operation(ctx, tx); err != nil {
				tm.logger.ErrorWithContext(ctx, "Batch operation failed", err, map[string]interface{}{
					"operation": op.Name,
					"index":     i,
				})

				// Rollback to savepoint
				if rbErr := savepoint.Rollback(ctx); rbErr != nil {
					return fmt.Errorf("operation %s failed: %w, savepoint rollback failed: %v", op.Name, err, rbErr)
				}

				errors = append(errors, fmt.Errorf("operation %s failed: %w", op.Name, err))

				// Call error handler if provided
				if op.OnError != nil {
					if handlerErr := op.OnError(ctx, err); handlerErr != nil {
						errors = append(errors, fmt.Errorf("error handler for %s failed: %w", op.Name, handlerErr))
					}
				}

				if !continueOnError {
					return errors[0]
				}
			} else {
				// Release savepoint on success
				if err := savepoint.Release(ctx); err != nil {
					tm.logger.WarnWithContext(ctx, "Failed to release savepoint", map[string]interface{}{
						"operation": op.Name,
						"error":     err.Error(),
					})
				}
			}
		}

		if len(errors) > 0 {
			tm.logger.WarnWithContext(ctx, "Batch execution completed with errors", map[string]interface{}{
				"total_operations":  len(operations),
				"failed_operations": len(errors),
			})
			return fmt.Errorf("batch execution had %d errors: %v", len(errors), errors)
		}

		tm.logger.InfoWithContext(ctx, "All batch operations completed successfully", map[string]interface{}{
			"total_operations": len(operations),
		})

		return nil
	})
}
