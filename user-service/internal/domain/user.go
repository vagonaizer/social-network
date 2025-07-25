package domain

import (
	"errors"
	"time"
)

// User - основная сущность пользователя (агрегат)
type User struct {
	id           int64
	firstName    string
	lastName     *string
	shortUrlName *string
	status       *string
	aboutMe      *string
	birthDate    *time.Time
	lastSeen     *time.Time
	createdAt    time.Time
	updatedAt    time.Time
	isOnline     bool
	location     *UserLocation
	settings     *UserSettings
}

// Конструктор
func NewUser(firstName string) (*User, error) {
	if firstName == "" {
		return nil, errors.New("firstName cannot be empty")
	}

	now := time.Now()
	return &User{
		firstName: firstName,
		createdAt: now,
		updatedAt: now,
		isOnline:  false,
		settings:  NewDefaultUserSettings(),
	}, nil
}

// Геттеры
func (u *User) ID() int64               { return u.id }
func (u *User) FirstName() string       { return u.firstName }
func (u *User) LastName() *string       { return u.lastName }
func (u *User) ShortUrlName() *string   { return u.shortUrlName }
func (u *User) Status() *string         { return u.status }
func (u *User) AboutMe() *string        { return u.aboutMe }
func (u *User) BirthDate() *time.Time   { return u.birthDate }
func (u *User) LastSeen() *time.Time    { return u.lastSeen }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) IsOnline() bool          { return u.isOnline }
func (u *User) Location() *UserLocation { return u.location }
func (u *User) Settings() *UserSettings { return u.settings }

// Сеттеры с валидацией
func (u *User) SetFirstName(firstName string) error {
	if firstName == "" {
		return errors.New("firstName cannot be empty")
	}
	u.firstName = firstName
	u.updatedAt = time.Now()
	return nil
}

func (u *User) SetLastName(lastName *string) {
	u.lastName = lastName
	u.updatedAt = time.Now()
}

func (u *User) SetShortUrlName(shortUrlName *string) error {
	// Можно добавить валидацию на уникальность и формат
	u.shortUrlName = shortUrlName
	u.updatedAt = time.Now()
	return nil
}

func (u *User) SetStatus(status *string) {
	u.status = status
	u.updatedAt = time.Now()
}

func (u *User) SetAboutMe(aboutMe *string) {
	u.aboutMe = aboutMe
	u.updatedAt = time.Now()
}

func (u *User) SetBirthDate(birthDate *time.Time) {
	u.birthDate = birthDate
	u.updatedAt = time.Now()
}

func (u *User) SetOnline() {
	u.isOnline = true
	u.lastSeen = nil // если онлайн, то lastSeen не нужен
	u.updatedAt = time.Now()
}

func (u *User) SetOffline() {
	u.isOnline = false
	now := time.Now()
	u.lastSeen = &now
	u.updatedAt = now
}

func (u *User) SetLocation(location *UserLocation) {
	u.location = location
	u.updatedAt = time.Now()
}

func (u *User) UpdateSettings(settings *UserSettings) {
	u.settings = settings
	u.updatedAt = time.Now()
}

// Бизнес-методы
func (u *User) GetFullName() string {
	if u.lastName != nil {
		return u.firstName + " " + *u.lastName
	}
	return u.firstName
}

func (u *User) IsProfileComplete() bool {
	return u.lastName != nil && u.birthDate != nil && u.aboutMe != nil
}
