package mocks

import (
	"log"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func GetRedisClientMock() (*redis.Client, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}

	return redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	}), s
}
