package domain

import "errors"

// UserLocation - value object для местоположения
type UserLocation struct {
	country string
	city    string
}

// Конструктор
func NewUserLocation(country, city string) (*UserLocation, error) {
	if country == "" {
		return nil, errors.New("country cannot be empty")
	}
	if city == "" {
		return nil, errors.New("city cannot be empty")
	}

	return &UserLocation{
		country: country,
		city:    city,
	}, nil
}

// Геттеры
func (ul *UserLocation) Country() string { return ul.country }
func (ul *UserLocation) City() string    { return ul.city }

// Value object должен быть immutable, поэтому сеттеров нет
// Для изменения создается новый объект

func (ul *UserLocation) String() string {
	return ul.city + ", " + ul.country
}

func (ul *UserLocation) Equals(other *UserLocation) bool {
	if other == nil {
		return false
	}
	return ul.country == other.country && ul.city == other.city
}
