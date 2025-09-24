package models

import "database/sql"

type User struct {
	ID             string
	Email          string
	Password       string
	FailedAttempts int
	LockedUntil    sql.NullTime
}