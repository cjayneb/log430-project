package models

type UserContactInfo struct {
	UserID int    `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required"`
}
