package helpers

import "social-network/auth-service/internal/domain"

// RoleHierarchy определяет иерархию ролей
var RoleHierarchy = map[domain.UserRoleType]int{
	domain.RoleUser:      1,
	domain.RoleModerator: 2,
	domain.RoleAdmin:     3,
}

// HasHigherRole проверяет, имеет ли пользователь роль выше указанной
func HasHigherRole(userRole, requiredRole domain.UserRoleType) bool {
	userLevel, userExists := RoleHierarchy[userRole]
	requiredLevel, requiredExists := RoleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// GetRoleLevel возвращает уровень роли
func GetRoleLevel(role domain.UserRoleType) int {
	if level, exists := RoleHierarchy[role]; exists {
		return level
	}
	return 0
}

// IsValidRole проверяет, является ли роль валидной
func IsValidRole(role domain.UserRoleType) bool {
	_, exists := RoleHierarchy[role]
	return exists
}

// GetAllRoles возвращает все доступные роли
func GetAllRoles() []domain.UserRoleType {
	roles := make([]domain.UserRoleType, 0, len(RoleHierarchy))
	for role := range RoleHierarchy {
		roles = append(roles, role)
	}
	return roles
}

// CanAssignRole проверяет, может ли пользователь с одной ролью назначить другую роль
func CanAssignRole(assignerRole, targetRole domain.UserRoleType) bool {
	assignerLevel := GetRoleLevel(assignerRole)
	targetLevel := GetRoleLevel(targetRole)

	// Можно назначать роли только ниже своей
	return assignerLevel > targetLevel
}
