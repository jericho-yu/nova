package lock

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jericho-yu/nova/src/util/dict"
)

type (
	// MapLock 字典锁：一个锁的集合
	MapLock struct {
		locks *dict.AnyDict[string, *itemLock]
	}

	// 锁项：一个集合锁中的每一项，包含：锁状态、锁值、超时时间、定时器
	itemLock struct {
		inUse   bool
		val     any
		timeout time.Duration
		timer   *time.Timer
	}
)

var (
	onceMapLock sync.Once
	mapLockIns  *MapLock
	MapLockApp  MapLock
)

func (*MapLock) New() *MapLock { return NewMapLock() }

func (*MapLock) Once() *MapLock { return OnceMapLock() }

// NewMapLock 实例化：字典锁
//
//go:fix 推荐使用：New方法
func NewMapLock() *MapLock { return &MapLock{locks: dict.Make[string, *itemLock]()} }

// OnceMapLock 单例化：字典锁
//
//go:fix 推荐使用：Once方法
func OnceMapLock() *MapLock {
	onceMapLock.Do(func() { mapLockIns = &MapLock{locks: dict.Make[string, *itemLock]()} })

	return mapLockIns
}

// Set 创建锁
func (my *MapLock) Set(key string, val any) error {
	_, exists := my.locks.Get(key)
	if exists {
		return fmt.Errorf("锁[%s]已存在", key)
	} else {
		my.locks.Set(key, &itemLock{val: val})
	}

	return nil
}

// SetMany 批量创建锁
func (my *MapLock) SetMany(items map[string]any) error {
	for idx, item := range items {
		err := my.Set(idx, item)
		if err != nil {
			my.DestroyAll()
			return err
		}
	}

	return nil
}

// Release 显式锁释放方法
func (r *itemLock) Release() {
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.inUse = false
}

// Destroy 删除锁
func (my *MapLock) Destroy(key string) {
	if il, ok := my.locks.Get(key); ok {
		il.Release()
		my.locks.RemoveByKey(key) // 删除键值对，以便垃圾回收
	}
}

// DestroyAll 删除所有锁
func (my *MapLock) DestroyAll() {
	my.locks.Each(func(key string, value *itemLock) {
		my.Destroy(key)
	})
}

// Lock 获取锁
func (my *MapLock) Lock(key string, timeout time.Duration) (*itemLock, error) {
	if item, exists := my.locks.Get(key); !exists {
		return nil, fmt.Errorf("锁[%s]不存在", key)
	} else {
		if item.inUse {
			return nil, fmt.Errorf("锁[%s]被占用", key)
		}

		// 设置锁占用
		item.inUse = true

		// 设置超时时间
		if timeout > 0 {
			item.timeout = timeout
			item.timer = time.AfterFunc(timeout, func() {
				if il, ok := my.locks.Get(key); ok {
					if il.timer != nil {
						il.Release()
					}
				}
			})
		}

		return item, nil
	}
}

// Try 尝试获取锁
func (my *MapLock) Try(key string) error {
	if item, exist := my.locks.Get(key); !exist {
		return fmt.Errorf("锁[%s]不存在", key)
	} else {
		if item.inUse {
			return fmt.Errorf("锁[%s]被占用", key)
		}
		return nil
	}
}

func DemoMapLock() {
	k8sLinks := map[string]any{
		"k8s-a": &struct{}{},
		"k8s-b": &struct{}{},
		"k8s-c": &struct{}{},
	}

	// 获取字典锁对象
	ml := OnceMapLock()

	// 批量创建锁
	storeErr := ml.SetMany(k8sLinks)
	if storeErr != nil {
		// 处理err
		log.Fatalln(storeErr.Error())
	}

	// 检测锁
	tryErr := ml.Try("k8s-a")
	if tryErr != nil {
		// 处理err
		log.Fatalln(tryErr.Error())
	}

	// 获取锁
	lock, lockErr := ml.Lock("k8s-a", time.Second*10) // 10秒业务处理不完也会过期 设置为：0则为永不过期
	if lockErr != nil {
		log.Fatalln(lockErr.Error())
	}
	defer lock.Release()

	// 处理业务...
}
