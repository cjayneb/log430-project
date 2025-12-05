package ports

import (
	"brokerx/notification-service/models"
	"context"
)

type UserService interface {
	GetUserContactInfo(ctx context.Context, userID int) (models.UserContactInfo, error)
}
