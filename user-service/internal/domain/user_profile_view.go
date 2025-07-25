package domain

import "time"

// UserProfileView - представление профиля пользователя для отображения
// Содержит вычисляемые поля для UI
type UserProfileView struct {
	User               *User
	IsFriend           bool
	IsBlocked          bool
	CanSendMessage     bool
	MutualFriendsCount int
}

// Конструктор
func NewUserProfileView(user *User) *UserProfileView {
	return &UserProfileView{
		User: user,
	}
}

// Методы для установки вычисляемых полей
func (upv *UserProfileView) SetFriendshipStatus(isFriend bool) {
	upv.IsFriend = isFriend
}

func (upv *UserProfileView) SetBlockedStatus(isBlocked bool) {
	upv.IsBlocked = isBlocked
}

func (upv *UserProfileView) SetCanSendMessage(canSend bool) {
	upv.CanSendMessage = canSend
}

func (upv *UserProfileView) SetMutualFriendsCount(count int) {
	upv.MutualFriendsCount = count
}

// Методы для получения отфильтрованной информации на основе настроек приватности
func (upv *UserProfileView) GetVisibleBirthDate() *time.Time {
	if upv.User.Settings().ShowBirthDate() {
		return upv.User.BirthDate()
	}
	return nil
}

func (upv *UserProfileView) GetVisibleLastSeen() *time.Time {
	if upv.User.Settings().ShowLastSeen() {
		return upv.User.LastSeen()
	}
	return nil
}

func (upv *UserProfileView) GetVisibleOnlineStatus() bool {
	if upv.User.Settings().ShowOnlineStatus() {
		return upv.User.IsOnline()
	}
	return false
}
