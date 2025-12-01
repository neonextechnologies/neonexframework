package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TxManager manages database transactions
type TxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes a function within a transaction
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// BeginTx starts a new transaction
func (tm *TxManager) BeginTx(ctx context.Context) *gorm.DB {
	return tm.db.WithContext(ctx).Begin()
}

// Commit commits the transaction
func (tm *TxManager) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback rolls back the transaction
func (tm *TxManager) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// TransactionalOperation is a helper for transaction operations
type TransactionalOperation struct {
	tx *gorm.DB
}

// NewTransactionalOperation creates a new transactional operation
func NewTransactionalOperation(tx *gorm.DB) *TransactionalOperation {
	return &TransactionalOperation{tx: tx}
}

// Execute executes operations within the transaction
func (to *TransactionalOperation) Execute(operations ...func(tx *gorm.DB) error) error {
	for _, op := range operations {
		if err := op(to.tx); err != nil {
			return fmt.Errorf("transaction operation failed: %w", err)
		}
	}
	return nil
}

// GetTx returns the transaction
func (to *TransactionalOperation) GetTx() *gorm.DB {
	return to.tx
}
