package postgres

import (
	"context"
	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userAuthRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserAuthRepository(db *pgxpool.Pool) repository.UserAuthRepository {
	return &userAuthRepositoryImpl{db: db}
}

func (r *userAuthRepositoryImpl) Create(userAuth *domain.UserAuth) error {
	query := `
        INSERT INTO user_auth (id, user_id, password_hash, last_login_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(context.Background(), query,
		userAuth.ID(),
		userAuth.UserID(),
		userAuth.PasswordHash(),
		userAuth.LastLoginAt(),
		userAuth.CreatedAt(),
		userAuth.UpdatedAt(),
	)

	return err
}

func (r *userAuthRepositoryImpl) GetByUserID(userID uuid.UUID) (*domain.UserAuth, error) {
	query := `
        SELECT id, user_id, password_hash, last_login_at, created_at, updated_at
        FROM user_auth
        WHERE user_id = $1
    `

	row := r.db.QueryRow(context.Background(), query, userID)

	var id, userId uuid.UUID
	var passwordHash string
	var lastLoginAt *time.Time
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &userId, &passwordHash, &lastLoginAt, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserAuthNotFound
		}
		return nil, err
	}

	userAuth := domain.NewUserAuth(userId, passwordHash)
	userAuth.SetID(id)
	userAuth.SetLastLoginAt(lastLoginAt)
	userAuth.SetCreatedAt(createdAt)
	userAuth.SetUpdatedAt(updatedAt)

	return userAuth, nil
}

func (r *userAuthRepositoryImpl) Update(userAuth *domain.UserAuth) error {
	query := `
        UPDATE user_auth 
        SET password_hash = $2, last_login_at = $3, updated_at = $4
        WHERE user_id = $1
    `

	result, err := r.db.Exec(context.Background(), query,
		userAuth.UserID(),
		userAuth.PasswordHash(),
		userAuth.LastLoginAt(),
		userAuth.UpdatedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserAuthNotFound
	}

	return nil
}

func (r *userAuthRepositoryImpl) Delete(userID uuid.UUID) error {
	query := `DELETE FROM user_auth WHERE user_id = $1`

	result, err := r.db.Exec(context.Background(), query, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserAuthNotFound
	}

	return nil
}
