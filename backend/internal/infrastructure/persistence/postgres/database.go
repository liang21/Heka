package postgres

import (
	"context"
	"fmt"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/shared/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) shared.TransactionManager {
	return &transactionManager{db: db}
}

type txKey struct{}

func (tm *transactionManager) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

// DBOrTx returns the transaction if available in context, otherwise returns the db.
func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

// isUniqueConstraintError checks if the error is a unique constraint violation.
func isUniqueConstraintError(err error, constraintName string) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "unique constraint") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "23505") ||
		(constraintName != "" && contains(errStr, constraintName))
}

// containsIgnoreCase is a simple case-insensitive substring check.
func containsIgnoreCase(s, substr string) bool {
	return contains(s, substr)
}

// contains is a simple case-insensitive substring check.
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			suc := substr[j]
			// Convert to lowercase for ASCII comparison
			if sc >= 'A' && sc <= 'Z' {
				sc = sc + 32
			}
			if suc >= 'A' && suc <= 'Z' {
				suc = suc + 32
			}
			if sc != suc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
