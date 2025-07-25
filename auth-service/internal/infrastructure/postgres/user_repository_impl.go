package postgres

import (
	"context"
	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"time"
)

type userRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *domain.User) error {
	query := `
        INSERT INTO users (id, email, username, display_name, is_verified, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.Exec(context.Background(), query,
		user.ID(),
		user.Email(),
		user.Username(),
		user.DisplayName(),
		user.IsVerified(),
		user.IsActive(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)

	return err
}

func (r *userRepositoryImpl) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `
        SELECT id, email, username, display_name, is_verified, is_active, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	row := r.db.QueryRow(context.Background(), query, id)

	var userID uuid.UUID
	var email, username, displayName string
	var isVerified, isActive bool
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &email, &username, &displayName, &isVerified, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user := domain.NewUser(email, username, displayName)
	user.SetID(userID)
	user.SetVerified(isVerified)
	user.SetActive(isActive)
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}

func (r *userRepositoryImpl) GetByEmail(email string) (*domain.User, error) {
	query := `
        SELECT id, email, username, display_name, is_verified, is_active, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	row := r.db.QueryRow(context.Background(), query, email)

	var userID uuid.UUID
	var userEmail, username, displayName string
	var isVerified, isActive bool
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &userEmail, &username, &displayName, &isVerified, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user := domain.NewUser(userEmail, username, displayName)
	user.SetID(userID)
	user.SetVerified(isVerified)
	user.SetActive(isActive)
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}

func (r *userRepositoryImpl) GetByUsername(username string) (*domain.User, error) {
	query := `
        SELECT id, email, username, display_name, is_verified, is_active, created_at, updated_at
        FROM users
        WHERE username = $1
    `

	row := r.db.QueryRow(context.Background(), query, username)

	var userID uuid.UUID
	var email, userName, displayName string
	var isVerified, isActive bool
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &email, &userName, &displayName, &isVerified, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user := domain.NewUser(email, userName, displayName)
	user.SetID(userID)
	user.SetVerified(isVerified)
	user.SetActive(isActive)
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}

func (r *userRepositoryImpl) Update(user *domain.User) error {
	query := `
        UPDATE users 
        SET email = $2, username = $3, display_name = $4, is_verified = $5, is_active = $6, updated_at = $7
        WHERE id = $1
    `

	result, err := r.db.Exec(context.Background(), query,
		user.ID(),
		user.Email(),
		user.Username(),
		user.DisplayName(),
		user.IsVerified(),
		user.IsActive(),
		user.UpdatedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

func (r *userRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

func (r *userRepositoryImpl) ExistsByEmail(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(context.Background(), query, email).Scan(&exists)

	return exists, err
}

func (r *userRepositoryImpl) ExistsByUsername(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.QueryRow(context.Background(), query, username).Scan(&exists)

	return exists, err
}

//  func scanUser(row pgx.Row) (*domain.User, error) {
// 	var userID uuid.UUID
// 	var email, username, displayName string
// 	var isVerified, isActive bool
// 	var createdAt, updatedAt time.Time

// 	err := row.Scan(&userID, &email, &username, &displayName, &isVerified, &isActive, &createdAt, &updatedAt)
// 	if err != nil {
// 		return nil, err
// 	}

// 	user := domain.NewUser(email, username, displayName)
// 	user.SetID(userID)
// 	user.SetVerified(isVerified)
// 	user.SetActive(isActive)
// 	user.SetCreatedAt(createdAt)
// 	user.SetUpdatedAt(updatedAt)

// 	return user, nil
// }
