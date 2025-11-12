package ports

import (
	"brokerx/user-service/models"
	"context"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Create(ctx context.Context, user *models.User) error
}
