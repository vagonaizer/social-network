package domain

// UserSettings - настройки пользователя
type UserSettings struct {
	userID           int64
	isPrivateProfile bool
	allowMessages    bool
	showOnlineStatus bool
	showLastSeen     bool
	showBirthDate    bool
}

// Конструктор с настройками по умолчанию
func NewDefaultUserSettings() *UserSettings {
	return &UserSettings{
		isPrivateProfile: false,
		allowMessages:    true,
		showOnlineStatus: true,
		showLastSeen:     true,
		showBirthDate:    true,
	}
}

func NewUserSettings(userID int64) *UserSettings {
	settings := NewDefaultUserSettings()
	settings.userID = userID
	return settings
}

// Геттеры
func (us *UserSettings) UserID() int64          { return us.userID }
func (us *UserSettings) IsPrivateProfile() bool { return us.isPrivateProfile }
func (us *UserSettings) AllowMessages() bool    { return us.allowMessages }
func (us *UserSettings) ShowOnlineStatus() bool { return us.showOnlineStatus }
func (us *UserSettings) ShowLastSeen() bool     { return us.showLastSeen }
func (us *UserSettings) ShowBirthDate() bool    { return us.showBirthDate }

// Сеттеры
func (us *UserSettings) SetPrivateProfile(isPrivate bool) {
	us.isPrivateProfile = isPrivate
}

func (us *UserSettings) SetAllowMessages(allow bool) {
	us.allowMessages = allow
}

func (us *UserSettings) SetShowOnlineStatus(show bool) {
	us.showOnlineStatus = show
}

func (us *UserSettings) SetShowLastSeen(show bool) {
	us.showLastSeen = show
}

func (us *UserSettings) SetShowBirthDate(show bool) {
	us.showBirthDate = show
}
