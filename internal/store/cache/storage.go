package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/supremed3v/social-media/internal/store"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
		Delete(context.Context, int64)
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}
}
