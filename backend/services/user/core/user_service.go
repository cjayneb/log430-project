package core

import (
	"brokerx/user-service/models"
	"brokerx/user-service/ports"
	"brokerx/user-service/util"
	"context"
)

type UserService interface {
	GetUserContactInfo(ctx context.Context, userId int) (models.User, error)
}

type UserServiceImpl struct {
	Repo ports.UserRepository
}

func (u *UserServiceImpl) GetUserContactInfo(ctx context.Context, userId int) (models.User, error) {
	log := util.FromContext(ctx)
	returnedUser, err := u.Repo.FindByUserId(ctx, userId)
	if err != nil {
		log.Error("error when fetching user", "error", err)
		return models.User{}, err
	}
	return *returnedUser, nil
}

var _ UserService = (*UserServiceImpl)(nil) // Ensure interface is implemented at compile time
