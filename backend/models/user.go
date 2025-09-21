package models

type User struct {
	ID       string
	Email    string `json:"email"`
	Password string `json:"password"`
}