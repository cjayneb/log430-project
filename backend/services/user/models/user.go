package models

import "database/sql"

type User struct {
	ID             int
	Email          string `json:"email" validate:"required"`
	Password       string `json:"password" validate:"required"`
	FailedAttempts int
	LockedUntil    sql.NullTime
}
