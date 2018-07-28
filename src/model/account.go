package model

type Account struct {
	AccountNumber int `gorm:"primary_key;auto_increment:false"`
	Pin           int
	Balance       int `json:"Balance,omitempty"`
	Overdraft     int
}
