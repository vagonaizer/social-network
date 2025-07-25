package postgres

import (
	"context"
	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRoleRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserRoleRepository(db *pgxpool.Pool) repository.UserRoleRepository {
	return &userRoleRepositoryImpl{db: db}
}

func (r *userRoleRepositoryImpl) Create(userRole *domain.UserRole) error {
	query := `
        INSERT INTO user_roles (id, user_id, role, granted_at, is_active)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.db.Exec(context.Background(), query,
		userRole.ID(),
		userRole.UserID(),
		string(userRole.Role()),
		userRole.GrantedAt(),
		userRole.IsActive(),
	)

	return err
}

func (r *userRoleRepositoryImpl) GetByUserID(userID uuid.UUID) ([]*domain.UserRole, error) {
	query := `
        SELECT id, user_id, role, granted_at, is_active
        FROM user_roles
        WHERE user_id = $1
        ORDER BY granted_at DESC
    `

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []*domain.UserRole

	for rows.Next() {
		var id, userId uuid.UUID
		var roleStr string
		var grantedAt time.Time
		var isActive bool

		err := rows.Scan(&id, &userId, &roleStr, &grantedAt, &isActive)
		if err != nil {
			return nil, err
		}

		userRole := domain.NewUserRole(userId, domain.UserRoleType(roleStr))
		userRole.SetID(id)
		userRole.SetGrantedAt(grantedAt)
		userRole.SetActive(isActive)

		userRoles = append(userRoles, userRole)
	}

	return userRoles, nil
}

func (r *userRoleRepositoryImpl) Update(userRole *domain.UserRole) error {
	query := `
        UPDATE user_roles 
        SET role = $2, is_active = $3
        WHERE id = $1
    `

	result, err := r.db.Exec(context.Background(), query,
		userRole.ID(),
		string(userRole.Role()),
		userRole.IsActive(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserRoleNotFound
	}

	return nil
}

func (r *userRoleRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserRoleNotFound
	}

	return nil
}
