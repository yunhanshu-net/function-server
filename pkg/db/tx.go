package db

import (
	"context"
	"gorm.io/gorm"
)

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := WithContextTxDB(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
