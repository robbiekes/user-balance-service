package entity

// User - структура для хранения данных о пользователе
type User struct {
	Id       int    `json:"-" db:"id"`
	Username string `json:"username" db:"username" binding:"required"`
	Password string `json:"password" db:"password" binding:"required"`
}
