package retry

import (
	"context"
	"math/rand"
	"time"
)

type Retry struct {
	sleep time.Duration
	fn    func() error
	ctx   context.Context
}

var RetryApp Retry

// New 实例化：重试器
func (*Retry) New() *Retry { return &Retry{sleep: time.Second, fn: nil, ctx: context.TODO()} }

// SetSleep 设置重试间隔
func (my *Retry) SetSleep(sleep time.Duration) *Retry {
	my.sleep = sleep

	return my
}

// SetFn 设置重试方法
func (my *Retry) SetFn(fn func() error) *Retry {
	my.fn = fn

	return my
}

// SetCtx 设置上下文
func (my *Retry) SetCtx(ctx context.Context) *Retry {
	my.ctx = ctx

	return my
}

// Do 指数退避
func (my *Retry) Do(attempts int) error {
	if my.fn == nil {
		return nil
	}

	if err := my.fn(); err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(my.sleep)
			return my.SetSleep(2 * my.sleep).Do(attempts)
		}
		return err
	}

	return nil
}

// WithContext 带上下文的重试
func (my *Retry) WithContext(attempts int) error {
	if my.fn == nil {
		return nil
	}

	if err := my.fn(); err != nil {
		if attempts--; attempts > 0 {
			select {
			case <-time.After(my.sleep):
				return my.SetSleep(2 * my.sleep).WithContext(attempts) // 指数退避
			case <-my.ctx.Done():
				return my.ctx.Err()
			}
		}
		return err
	}

	return nil
}

func (my *Retry) WithContextAndJitter(attempts int) error {
	if my.fn == nil {
		return nil
	}

	if err := my.fn(); err != nil {
		if attempts--; attempts > 0 {
			// 加入随机退避
			jitter := time.Duration(rand.Int63n(int64(my.sleep)))
			my.sleep = my.sleep + jitter

			select {
			case <-time.After(my.sleep):
				return my.SetSleep(2 * my.sleep).WithContextAndJitter(attempts) // 指数退避
			case <-my.ctx.Done():
				return my.ctx.Err()
			}
		}
		return err
	}

	return nil
}
