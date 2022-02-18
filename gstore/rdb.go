package gstore

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

const (
	defaultRedisPoolMinIdle = 2 // 连接池空闲连接数量
)

// InitRdb 初始化redis
func InitRdb(cfg redis.Options) *redis.Client {
	return connRdb(cfg)
}

func connRdb(options redis.Options) *redis.Client {
	if options.MinIdleConns == 0 {
		options.MinIdleConns = defaultRedisPoolMinIdle
	}
	Rdb := redis.NewClient(&options)
	_, err := Rdb.Ping(context.Background()).Result()
	if err == redis.Nil {
		log.Printf("[app.gstore] Nil reply returned by Rdb when key does not exist.")
	} else if err != nil {
		log.Printf("[app.gstore] redis fail, err=%s", err)
		panic(err)
	} else {
		log.Printf("[app.gstore] redis success")
	}
	return Rdb
}

// RdbSurvive redis存活检测
func RdbSurvive(rdb *redis.Client) error {
	err := rdb.Ping(context.Background()).Err()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
