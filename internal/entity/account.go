package entity

type Account struct {
	Id      int `json:"id" db:"id" binding:"required"`
	Balance int `json:"balance" db:"balance" binding:"required"`
}
