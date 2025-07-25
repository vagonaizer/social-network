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

type refreshTokenRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) repository.RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{db: db}
}

func (r *refreshTokenRepositoryImpl) Create(token *domain.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (id, user_id, token, expires_at, is_revoked, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(context.Background(), query,
		token.ID(),
		token.UserID(),
		token.Token(),
		token.ExpiresAt(),
		token.IsRevoked(),
		token.CreatedAt(),
	)

	return err
}

func (r *refreshTokenRepositoryImpl) GetByToken(tokenStr string) (*domain.RefreshToken, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_revoked, created_at
        FROM refresh_tokens
        WHERE token = $1
    `

	row := r.db.QueryRow(context.Background(), query, tokenStr)

	var id, userID uuid.UUID
	var token string
	var expiresAt, createdAt time.Time
	var isRevoked bool

	err := row.Scan(&id, &userID, &token, &expiresAt, &isRevoked, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrRefreshTokenNotFound
		}
		return nil, err
	}

	refreshToken := domain.NewRefreshToken(userID, token, expiresAt)
	refreshToken.SetID(id)
	refreshToken.SetRevoked(isRevoked)
	refreshToken.SetCreatedAt(createdAt)

	return refreshToken, nil
}

func (r *refreshTokenRepositoryImpl) GetByUserID(userID uuid.UUID) ([]*domain.RefreshToken, error) {
	query := `
        SELECT id, user_id, token, expires_at, is_revoked, created_at
        FROM refresh_tokens
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken

	for rows.Next() {
		var id, userId uuid.UUID
		var token string
		var expiresAt, createdAt time.Time
		var isRevoked bool

		err := rows.Scan(&id, &userId, &token, &expiresAt, &isRevoked, &createdAt)
		if err != nil {
			return nil, err
		}

		refreshToken := domain.NewRefreshToken(userId, token, expiresAt)
		refreshToken.SetID(id)
		refreshToken.SetRevoked(isRevoked)
		refreshToken.SetCreatedAt(createdAt)

		tokens = append(tokens, refreshToken)
	}

	return tokens, nil
}

func (r *refreshTokenRepositoryImpl) Update(token *domain.RefreshToken) error {
	query := `
        UPDATE refresh_tokens 
        SET is_revoked = $2
        WHERE id = $1
    `

	result, err := r.db.Exec(context.Background(), query, token.ID(), token.IsRevoked())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepositoryImpl) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(context.Background(), query, userID)
	return err
}
