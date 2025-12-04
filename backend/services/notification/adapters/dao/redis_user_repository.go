package dao_adapters

import (
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"brokerx/notification-service/util"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUserRepository struct {
	Rdb *redis.Client
}

func (r *RedisUserRepository) SetContactInfo(ctx context.Context, info models.UserContactInfo) error {
	log := util.FromContext(ctx)
	log.Info("Trying to set user contact info in redis")

	data, err := json.Marshal(info)
	if err != nil {
		log.Error("error when marshaling user contact info", "error", err)
		return err
	}

	key := fmt.Sprintf("usercontact:%d", info.UserID)
	expiration := 5 * time.Minute
	_, err = r.Rdb.Set(ctx, key, data, expiration).Result()
	if err != nil {
		log.Error("error when setting user contact info", "error", err)
		return err
	}

	return nil
}

func (r *RedisUserRepository) GetContactInfo(ctx context.Context, userId int) (models.UserContactInfo, error) {
	log := util.FromContext(ctx)
	log.Info("Trying to get user contact info from redis")

	key := fmt.Sprintf("usercontact:%d", userId)
	result, err := r.Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Info("user contact info is not in cache")
		return models.UserContactInfo{}, nil
	} else if err != nil {
		log.Error("error when fetching user contact info", "error", err)
		return models.UserContactInfo{}, err
	}

	var info models.UserContactInfo
	if err := json.Unmarshal([]byte(result), &info); err != nil {
		log.Error("error when unmarshaling user contact info", "error", err)
		return models.UserContactInfo{}, err
	}

	return info, nil
}

var _ ports.UserRepository = (*RedisUserRepository)(nil) // Ensure interface is implemented at compile time
