package redisPool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"nova/src/util/dict"

	rds "github.com/redis/go-redis/v9"
)

type (
	RedisPool struct {
		conns *dict.AnyDict[string, *redisConn]
	}

	redisConn struct {
		prefix string
		conn   *rds.Client
	}
)

var (
	redisPoolIns  *RedisPool
	redisPoolOnce sync.Once
	RedisPoolApp  RedisPool
)

func (*RedisPool) Once(redisSetting *RedisSetting) *RedisPool { return OnceRedisPool(redisSetting) }

// OnceRedisPool 单例化：redis 链接
//
//go:fix 推荐使用：Once方法
func OnceRedisPool(redisSetting *RedisSetting) *RedisPool {
	redisPoolOnce.Do(func() {
		redisPoolIns = &RedisPool{}
		redisPoolIns.conns = dict.Make[string, *redisConn]()

		if len(redisSetting.Pool) > 0 {
			for _, pool := range redisSetting.Pool {
				redisPoolIns.conns.Set(pool.Key, &redisConn{
					prefix: fmt.Sprintf("%s:%s", redisSetting.Prefix, pool.Prefix),
					conn: rds.NewClient(&rds.Options{
						Addr:     fmt.Sprintf("%s:%d", redisSetting.Host, redisSetting.Port),
						Password: redisSetting.Password,
						DB:       pool.DbNum,
					}),
				})
			}
		}
	})

	return redisPoolIns
}

// GetClient 获取链接和链接前缀
func (*RedisPool) GetClient(key string) (string, *rds.Client) {
	if client, exist := redisPoolIns.conns.Get(key); exist {
		return client.prefix, client.conn
	}

	return "", nil
}

// Get 获取值
func (*RedisPool) Get(clientName, key string) (string, error) {
	var (
		err         error
		prefix, ret string
		client      *rds.Client
	)

	prefix, client = redisPoolIns.GetClient(clientName)
	if client == nil {
		return "", fmt.Errorf("没有找到redis链接：%s", clientName)
	}

	ret, err = client.Get(context.Background(), fmt.Sprintf("%s:%s", prefix, key)).Result()
	if err != nil {
		if errors.Is(err, rds.Nil) {
			return "", nil
		} else {
			return "", err
		}
	}

	return ret, nil
}

// Set 设置值
func (*RedisPool) Set(clientName, key string, val any, exp time.Duration) (string, error) {
	var (
		prefix string
		client *rds.Client
	)

	prefix, client = redisPoolIns.GetClient(clientName)
	if client == nil {
		return "", fmt.Errorf("没有找到redis链接：%s", clientName)
	}

	return client.Set(context.Background(), fmt.Sprintf("%s:%s", prefix, key), val, exp).Result()
}

// Close 关闭链接
func (my *RedisPool) Close(key string) error {
	if client, exist := redisPoolIns.conns.Get(key); exist {
		return client.conn.Close()
	}

	return nil
}

// Clean 清理链接
func (*RedisPool) Clean() {
	for key, val := range redisPoolIns.conns.ToMap() {
		_ = val.conn.Close()
		redisPoolIns.conns.RemoveByKey(key)
	}
}
