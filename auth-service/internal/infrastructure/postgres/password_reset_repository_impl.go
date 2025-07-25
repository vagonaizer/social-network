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

type passwordResetRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepository(db *pgxpool.Pool) repository.PasswordResetRepository {
	return &passwordResetRepositoryImpl{db: db}
}

func (r *passwordResetRepositoryImpl) Create(reset *domain.PasswordReset) error {
	query := `
        INSERT INTO password_resets (id, user_id, token, expires_at, is_used, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(context.Background(), query,
		reset.ID(),
		reset.UserID(),
		reset.Token(),
		reset.ExpiresAt(),
		reset.IsUsed(),
		reset.CreatedAt(),
	)

	return err
}

func (r *passwordResetRepositoryImpl) GetByToken(token string) (*domain.PasswordReset, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_used, created_at
        FROM password_resets
        WHERE token = $1
    `

	row := r.db.QueryRow(context.Background(), query, token)

	var id, userID uuid.UUID
	var tokenStr string
	var expiresAt, createdAt time.Time
	var isUsed bool

	err := row.Scan(&id, &userID, &tokenStr, &expiresAt, &isUsed, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrPasswordResetNotFound
		}
		return nil, err
	}

	passwordReset := domain.NewPasswordReset(userID, tokenStr, expiresAt)
	passwordReset.SetID(id)
	passwordReset.SetUsed(isUsed)
	passwordReset.SetCreatedAt(createdAt)

	return passwordReset, nil
}

func (r *passwordResetRepositoryImpl) GetByUserID(userID uuid.UUID) (*domain.PasswordReset, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_used, created_at
        FROM password_resets
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `

	row := r.db.QueryRow(context.Background(), query, userID)

	var id, userId uuid.UUID
	var token string
	var expiresAt, createdAt time.Time
	var isUsed bool

	err := row.Scan(&id, &userId, &token, &expiresAt, &isUsed, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrPasswordResetNotFound
		}
		return nil, err
	}

	passwordReset := domain.NewPasswordReset(userId, token, expiresAt)
	passwordReset.SetID(id)
	passwordReset.SetUsed(isUsed)
	passwordReset.SetCreatedAt(createdAt)

	return passwordReset, nil
}

func (r *passwordResetRepositoryImpl) Update(reset *domain.PasswordReset) error {
	query := `
        UPDATE password_resets 
        SET is_used = $2
        WHERE id = $1
    `

	result, err := r.db.Exec(context.Background(), query, reset.ID(), reset.IsUsed())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrPasswordResetNotFound
	}

	return nil
}

func (r *passwordResetRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM password_resets WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrPasswordResetNotFound
	}

	return nil
}
