package models

type NotificationPreference struct {
	ID     int
	UserID int
	Email  bool
	SMS    bool
	Push   bool
}
