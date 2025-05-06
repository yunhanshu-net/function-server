package db

import (
	"context"
	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "gorm_tx_db"

func WithContextTxDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, db)
}

func GetContextDB(ctx context.Context, defaultDb *gorm.DB) *gorm.DB {
	value := ctx.Value(txKey)
	if value == nil {
		return defaultDb
	}
	vm, ok := value.(*gorm.DB)
	if !ok {
		return defaultDb
	}
	return vm
}
