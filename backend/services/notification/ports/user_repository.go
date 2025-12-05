package ports

import (
	"brokerx/notification-service/models"
	"context"
)

type UserRepository interface {
	GetContactInfo(ctx context.Context, userId int) (models.UserContactInfo, error)
	SetContactInfo(ctx context.Context, info models.UserContactInfo) error
}
