package rabbit

import (
	"sync"

	"nova/src/util/dict"
)

type (
	Pool struct {
		rabbitConns *dict.AnyDict[string, *Rabbit]
	}
)

var (
	PoolApp  Pool
	poolOnce sync.Once
	poolIns  *Pool
)

// Once 单例化：rabbit-mq连接池
func (*Pool) Once() *Pool {
	poolOnce.Do(func() {
		poolIns = &Pool{rabbitConns: dict.Make[string, *Rabbit]()}
	})

	return poolIns
}

// Set 添加链接
func (*Pool) Set(key string, value *Rabbit) *Pool {
	poolIns.rabbitConns.Set(key, value)
	return poolIns
}

// Get 获取链接
func (*Pool) Get(key string) *Rabbit {
	val, _ := poolIns.rabbitConns.Get(key)

	return val
}
