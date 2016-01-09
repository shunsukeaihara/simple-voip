package voip

import (
	"errors"
	"strconv"
	"strings"

	"gopkg.in/redis.v3"
)

type RedisCli struct {
	*redis.Client
}

func NewRedisClient(addr string, dbNum int) *RedisCli {
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   int64(dbNum),
	})
	return &RedisCli{client}
}

func (r *RedisCli) CheckJoinKey(key []byte) (int, int, error) {
	val, err := r.Get(string(key)).Result()
	if err != nil {
		return 0, 0, err
	} else {
		l := strings.Split(val, ",")
		if len(l) != 2 {
			return 0, 0, errors.New("join key's value is broken")
		}
		userId, err := strconv.Atoi(l[0])
		if err != nil {
			return 0, 0, errors.New("join key's value is broken")
		}
		roomId, err := strconv.Atoi(l[1])
		if err != nil {
			return 0, 0, errors.New("join key's value is broken")
		}
		return userId, roomId, nil
	}
}
