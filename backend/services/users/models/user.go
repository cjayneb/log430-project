package models

import "database/sql"

type User struct {
	ID             int
	Email          string
	Password       string
	FailedAttempts int
	LockedUntil    sql.NullTime
}
