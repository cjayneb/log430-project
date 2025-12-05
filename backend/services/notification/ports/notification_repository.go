package ports

import (
	"brokerx/notification-service/models"
	"context"
)

type NotificationRepository interface {
	FindByUserId(ctx context.Context, userId int) (*models.NotificationPreference, error)
	Update(ctx context.Context, preference models.NotificationPreference) error
	Create(ctx context.Context, preference *models.NotificationPreference) error
}
