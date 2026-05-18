package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
