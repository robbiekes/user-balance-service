package entity

// User - структура для заполнения данных о пользователе
type User struct {
	Id       int    `json:"-" db:"id"`
	Username string `json:"username" db:"username" validate:"required"`
	Password string `json:"password" db:"password" validate:"required"`
}
