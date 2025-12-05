package models

import "database/sql"

type User struct {
	ID             int    `json:"user_id"`
	Email          string `json:"email" validate:"required"`
	Password       string `json:"password" validate:"required"`
	FirstName      string        `json:"first_name" validate:"required"`
    LastName       string        `json:"last_name" validate:"required"`
    Status         string        `json:"status"`
	FailedAttempts int
	LockedUntil    sql.NullTime
}
