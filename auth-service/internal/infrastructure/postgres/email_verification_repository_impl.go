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

type emailVerificationRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) repository.EmailVerificationRepository {
	return &emailVerificationRepositoryImpl{db: db}
}

func (r *emailVerificationRepositoryImpl) Create(verification *domain.EmailVerification) error {
	query := `
        INSERT INTO email_verifications (id, user_id, token, expires_at, is_used, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(context.Background(), query,
		verification.ID(),
		verification.UserID(),
		verification.Token(),
		verification.ExpiresAt(),
		verification.IsUsed(),
		verification.CreatedAt(),
	)

	return err
}

func (r *emailVerificationRepositoryImpl) GetByToken(token string) (*domain.EmailVerification, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_used, created_at
        FROM email_verifications
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
			return nil, repository.ErrEmailVerificationNotFound
		}
		return nil, err
	}

	verification := domain.NewEmailVerification(userID, tokenStr, expiresAt)
	verification.SetID(id)
	verification.SetUsed(isUsed)
	verification.SetCreatedAt(createdAt)

	return verification, nil
}

func (r *emailVerificationRepositoryImpl) GetByUserID(userID uuid.UUID) (*domain.EmailVerification, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_used, created_at
        FROM email_verifications
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
			return nil, repository.ErrEmailVerificationNotFound
		}
		return nil, err
	}

	verification := domain.NewEmailVerification(userId, token, expiresAt)
	verification.SetID(id)
	verification.SetUsed(isUsed)
	verification.SetCreatedAt(createdAt)

	return verification, nil
}

func (r *emailVerificationRepositoryImpl) Update(verification *domain.EmailVerification) error {
	query := `
        UPDATE email_verifications 
        SET is_used = $2
        WHERE id = $1
    `

	result, err := r.db.Exec(context.Background(), query, verification.ID(), verification.IsUsed())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrEmailVerificationNotFound
	}

	return nil
}

func (r *emailVerificationRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM email_verifications WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrEmailVerificationNotFound
	}

	return nil
}
