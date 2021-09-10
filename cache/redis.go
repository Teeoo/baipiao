package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	Config "github.com/teeoo/baipiao/config"
	"log"
	"strconv"
)

var Redis *redis.Client

func init() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", Config.Config.Redis.Addr, strconv.FormatInt(int64(Config.Config.Redis.Port), 10)),
		Password: Config.Config.Redis.Password,
		DB:       1,
	})
	pong, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Println("redis 连接失败：", pong, err)
		return
	}
	log.Println("redis 连接成功:", pong)
}
