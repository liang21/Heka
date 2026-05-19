// tasks.md: T044 | TDD GREEN Phase for UserRepository
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
	"gorm.io/gorm"
)

// UserModel is the GORM model for users table
type UserModel struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email_deleted"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
	DeletedAt    *time.Time `gorm:"index:idx_users_email_deleted"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

// UserRepository implements user.UserRepository using GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	db := DBOrTx(ctx, r.db)
	model := r.domainToModel(u)

	if err := db.Create(model).Error; err != nil {
		// Check for unique constraint violation (duplicate email)
		if isUniqueConstraintError(err, "idx_users_email_deleted") {
			return shared.ErrUserEmailExists
		}
		// Check for not null constraint violation (empty email)
		errMsg := err.Error()
		if containsIgnoreCase(errMsg, "23502") || containsIgnoreCase(errMsg, "null constraint") || containsIgnoreCase(errMsg, "cannot be null") {
			return shared.ErrSysValidation
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	db := DBOrTx(ctx, r.db)
	var model UserModel

	err := db.Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, shared.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	db := DBOrTx(ctx, r.db)
	var model UserModel

	err := db.Where("email = ? AND deleted_at IS NULL", email).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, shared.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// Update updates a user in the database
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	db := DBOrTx(ctx, r.db)
	model := r.domainToModel(u)

	result := db.Model(&UserModel{}).
		Where("id = ? AND deleted_at IS NULL", string(model.ID)).
		Updates(map[string]interface{}{
			"name":          model.Name,
			"email":         model.Email,
			"password_hash": model.PasswordHash,
			"updated_at":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}

	return nil
}

// SoftDelete soft deletes a user by ID
func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	db := DBOrTx(ctx, r.db)
	now := time.Now()

	result := db.Model(&UserModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}

	return nil
}

// domainToModel converts domain entity to GORM model
func (r *UserRepository) domainToModel(u *user.User) *UserModel {
	return &UserModel{
		ID:           string(u.ID),
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		DeletedAt:    u.DeletedAt,
	}
}

// modelToDomain converts GORM model to domain entity
func (r *UserRepository) modelToDomain(model *UserModel) *user.User {
	return &user.User{
		ID:           shared.ID(model.ID),
		Name:         model.Name,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		DeletedAt:    model.DeletedAt,
	}
}
