package dao_adapters

import (
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"context"
	"database/sql"
)

type SQLNotificationRepository struct {
	DB *sql.DB
}

func (repo *SQLNotificationRepository) Create(ctx context.Context, preference *models.NotificationPreference) error {
	query := `
        INSERT INTO notification_preference (user_id, email, sms, push)
        VALUES (?, ?, ?, ?)
    `
	_, err := repo.DB.ExecContext(ctx, query, preference.UserID, preference.Email, preference.SMS, preference.Push)
	return err
}

func (repo *SQLNotificationRepository) FindByUserId(ctx context.Context, userId int) (*models.NotificationPreference, error) {
	row := repo.DB.QueryRowContext(ctx, "SELECT email, sms, push FROM notification_preference WHERE user_id=?", userId)

	var pref models.NotificationPreference
	e := row.Scan(&pref.Email, &pref.SMS, &pref.Push)
	if e == sql.ErrNoRows {
		pref.UserID = userId
		_ = repo.Create(ctx, &pref)
		return &pref, nil
	}
	if e != nil {
		return nil, e
	}

	return &pref, nil
}

func (repo *SQLNotificationRepository) Update(ctx context.Context, preference models.NotificationPreference) error {
	_, e := repo.DB.Exec(
		"UPDATE notification_preference SET email=?, sms=?, push=? WHERE user_id=?",
		preference.Email,
		preference.SMS,
		preference.Push,
		preference.UserID,
	)
	return e
}

var _ ports.NotificationRepository = (*SQLNotificationRepository)(nil) // Ensure interface is implemented at compile time
