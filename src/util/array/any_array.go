package array

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"nova/src/util/operation"
)

type (
	AnyArray[T any] struct {
		data []T
		mu   sync.RWMutex
	}
)

// New 实例化
func New[T any](list []T) *AnyArray[T] { return &AnyArray[T]{data: list, mu: sync.RWMutex{}} }

// NewDestruction 通过解构参数实例化
func NewDestruction[T any](list ...T) *AnyArray[T] {
	return &AnyArray[T]{data: list, mu: sync.RWMutex{}}
}

// Make 初始化
func Make[T any](size int) *AnyArray[T] {
	return &AnyArray[T]{data: make([]T, size), mu: sync.RWMutex{}}
}

// Lock 加锁：写锁
func (my *AnyArray[T]) Lock() *AnyArray[T] {
	my.mu.Lock()
	return my
}

// Unlock 解锁：写锁
func (my *AnyArray[T]) Unlock() *AnyArray[T] {
	my.mu.Unlock()
	return my
}

// RLock 加锁：读锁
func (my *AnyArray[T]) RLock() *AnyArray[T] {
	my.mu.RLock()
	return my
}

// RUnlock 解锁：读锁
func (my *AnyArray[T]) RUnlock() *AnyArray[T] {
	my.mu.RUnlock()
	return my
}

// isEmpty 判断是否为空
func (my *AnyArray[T]) isEmpty() bool { return len(my.data) == 0 }

// IsEmpty 判断是否为空
func (my *AnyArray[T]) IsEmpty() bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.isEmpty()
}

// IsNotEmpty 判断是否不为空
func (my *AnyArray[T]) IsNotEmpty() bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return !my.isEmpty()
}

// has 检查key是否存在
func (my *AnyArray[T]) has(k int) bool { return k >= 0 && k < len(my.data) }

// Has 检查是否存在
func (my *AnyArray[T]) Has(k int) bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.has(k)
}

func (my *AnyArray[T]) set(k int, v T) *AnyArray[T] {
	my.data[k] = v
	return my
}

// Set 设置值
func (my *AnyArray[T]) Set(k int, v T) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.set(k, v)
}

func (my *AnyArray[T]) get(idx int) T { return my.data[idx] }

// Get 获取值
func (my *AnyArray[T]) Get(idx int) T {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.get(idx)
}

func (my *AnyArray[T]) getByIndexes(indexes ...int) []T {
	res := make([]T, len(indexes))

	for k, idx := range indexes {
		res[k] = my.data[idx]
	}

	return res
}

// GetByIndexes 通过多索引获取内容
func (my *AnyArray[T]) GetByIndexes(indexes ...int) *AnyArray[T] {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return New(my.getByIndexes(indexes...))
}

func (my *AnyArray[T]) append(v ...T) *AnyArray[T] {
	my.data = append(my.data, v...)

	return my
}

// Append 追加
func (my *AnyArray[T]) Append(v ...T) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.append(v...)
}

func (my *AnyArray[T]) first() T { return my.data[0] }

// First 获取第一个值
func (my *AnyArray[T]) First() T {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.first()
}

func (my *AnyArray[T]) last() T {
	var t T

	return operation.Ternary(my.Len() > 1, my.data[len(my.data)-1], operation.Ternary(my.Len() == 0, t, my.data[0]))
}

// Last 获取最后一个值
func (my *AnyArray[T]) Last() T {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.last()
}

func (my *AnyArray[T]) toSlice() []T {
	var ret = make([]T, len(my.data))
	copy(ret, my.data)

	return ret
}

// ToSlice 获取全部值：到切片
func (my *AnyArray[T]) ToSlice() []T {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.toSlice()
}

func (my *AnyArray[T]) getIndexes() []int {
	var indexes = make([]int, len(my.data))
	for i := range my.data {
		indexes[i] = i
	}

	return indexes
}

// GetIndexes 获取所有索引
func (my *AnyArray[T]) GetIndexes() []int {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.getIndexes()
}

func (my *AnyArray[T]) getIndexByValue(value T) int {
	for idx, val := range my.data {
		if reflect.DeepEqual(val, value) {
			return idx
		}
	}

	return -1
}

// GetIndexByValue 根据值获取索引下标
func (my *AnyArray[T]) GetIndexByValue(value T) int {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.getIndexByValue(value)
}

func (my *AnyArray[T]) getIndexesByValues(values ...T) []int {
	var indexes []int
	for _, value := range values {
		for idx, val := range my.data {
			if reflect.DeepEqual(val, value) {
				indexes = append(indexes, idx)
			}
		}
	}

	return indexes
}

// GetIndexesByValues 通过值获取索引下标
func (my *AnyArray[T]) GetIndexesByValues(values ...T) *AnyArray[int] {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return New(my.getIndexesByValues(values...))
}

func (my *AnyArray[T]) copy() *AnyArray[T] {
	return New(my.data)
}

// Copy 复制自己
func (my *AnyArray[T]) Copy() *AnyArray[T] {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.copy()
}

func (my *AnyArray[T]) shuffle() *AnyArray[T] {
	randStr := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range my.data {
		j := randStr.Intn(i + 1)                        // 生成 [0, i] 范围内的随机数
		my.data[i], my.data[j] = my.data[j], my.data[i] // 交换元素
	}

	return my
}

// Shuffle 打乱切片中的元素顺序
func (my *AnyArray[T]) Shuffle() *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.shuffle()
}

func (my *AnyArray[T]) len() int {
	return len(my.data)
}

// Len 获取数组长度
func (my *AnyArray[T]) Len() int {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.len()
}

func (my *AnyArray[T]) lenWithoutEmpty() int { return my.copy().removeEmpty().len() }

// LenWithoutEmpty 获取非0值长度
func (my *AnyArray[T]) LenWithoutEmpty() int {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.lenWithoutEmpty()
}

func (my *AnyArray[T]) filter(fn func(v T) bool) *AnyArray[T] {
	j := 0
	ret := make([]T, len(my.data))
	for i := range my.data {
		if fn(my.data[i]) {
			ret[j] = my.data[i]
			j++
		}
	}

	my.data = ret[:j]
	return my
}

// Filter 过滤数组值
func (my *AnyArray[T]) Filter(fn func(v T) bool) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.filter(fn)
}

func (my *AnyArray[T]) removeEmpty() *AnyArray[T] {
	var data = make([]T, 0)

	for _, item := range my.data {
		ref := reflect.ValueOf(item)

		if ref.Kind() == reflect.Ptr {
			if ref.IsNil() {
				continue
			}
			if ref.Elem().IsZero() {
				continue
			}
		} else {
			if ref.IsZero() {
				continue
			}
		}

		data = append(data, item)
	}

	return New(data)
}

// RemoveEmpty 清除0值元素
func (my *AnyArray[T]) RemoveEmpty() *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.removeEmpty()
}

func (my *AnyArray[T]) join(sep string) string {
	values := make([]string, my.len())
	for idx, datum := range my.data {
		values[idx] = fmt.Sprintf("%v", datum)
	}
	return strings.Join(values, sep)
}

// Join 拼接字符串
func (my *AnyArray[T]) Join(sep string) string {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.join(sep)
}

func (my *AnyArray[T]) joinWithoutEmpty(sep string) string {
	values := make([]string, my.copy().removeEmpty().len())
	j := 0
	for _, datum := range my.copy().removeEmpty().toSlice() {
		values[j] = fmt.Sprintf("%v", datum)
		j++
	}

	return strings.Join(values, sep)
}

// JoinWithoutEmpty 拼接非空元素
func (my *AnyArray[T]) JoinWithoutEmpty(seps ...string) string {
	my.mu.Lock()
	defer my.mu.Unlock()

	var sep = " "
	if len(seps) > 0 {
		sep = seps[0]
	}

	return my.joinWithoutEmpty(sep)
}

func (my *AnyArray[T]) in(target T) bool {
	for _, element := range my.data {
		if reflect.DeepEqual(target, element) {
			return true
		}
	}

	return false
}

// In 检查值是否存在
func (my *AnyArray[T]) In(target T) bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.in(target)
}

func (my *AnyArray[T]) notIn(target T) bool { return !my.in(target) }

// NotIn 检查值是否不存在
func (my *AnyArray[T]) NotIn(target T) bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.notIn(target)
}

func (my *AnyArray[T]) allEmpty() bool { return my.copy().removeEmpty().len() == 0 }

// AllEmpty 判断当前数组是否0空
func (my *AnyArray[T]) AllEmpty() bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.allEmpty()
}

func (my *AnyArray[T]) anyEmpty() bool { return my.copy().removeEmpty().len() != len(my.data) }

// AnyEmpty 判断当前数组中是否存在0值
func (my *AnyArray[T]) AnyEmpty() bool {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.anyEmpty()
}

func (my *AnyArray[T]) chunk(size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(my.data); i += size {
		end := i + size
		if end > len(my.data) {
			end = len(my.data)
		}
		chunks = append(chunks, my.data[i:end])
	}

	return chunks
}

// Chunk 分块
func (my *AnyArray[T]) Chunk(size int) [][]T {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.chunk(size)
}

func (my *AnyArray[T]) pluck(fn func(item T) any) *AnyArray[any] {
	var ret = make([]any, 0)
	for _, v := range my.data {
		ret = append(ret, fn(v))
	}

	return New(ret)
}

// Pluck 获取数组中指定字段的值
func (my *AnyArray[T]) Pluck(fn func(item T) any) *AnyArray[any] {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.pluck(fn)
}

func (my *AnyArray[T]) unique() *AnyArray[T] {
	seen := make(map[string]struct{}) // 使用空结构体作为值，因为我们只关心键
	result := make([]T, 0)

	for _, value := range my.data {
		key := fmt.Sprintf("%v", value)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, value)
		}
	}

	my.data = result

	return my
}

// Unique 切片去重
func (my *AnyArray[T]) Unique() *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.unique()
}

func (my *AnyArray[T]) removeByIndex(index int) *AnyArray[T] {
	if index < 0 || index >= len(my.data) {
		return my
	}

	my.data = append(my.data[:index], my.data[index+1:]...)
	return my
}

// RemoveByIndex 根据索引删除元素
func (my *AnyArray[T]) RemoveByIndex(index int) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.removeByIndex(index)
}

func (my *AnyArray[T]) removeByIndexes(indexes ...int) *AnyArray[T] {
	for _, index := range indexes {
		my.removeByIndex(index)
	}

	return my
}

// RemoveByIndexes 根据索引删除元素
func (my *AnyArray[T]) RemoveByIndexes(indexes ...int) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.removeByIndexes(indexes...)
}

func (my *AnyArray[T]) removeByValue(target T) *AnyArray[T] {
	var ret = make([]T, len(my.data))
	j := 0
	for _, value := range my.data {
		if !reflect.DeepEqual(value, target) {
			ret[j] = value
			j++
		}
	}
	my.data = ret[:j]

	return my
}

// RemoveByValue 删除数组中对应的目标
func (my *AnyArray[T]) RemoveByValue(target T) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.removeByValue(target)
}

func (my *AnyArray[T]) removeByValues(targets ...T) *AnyArray[T] {
	for _, target := range targets {
		my.removeByValue(target)
	}

	return my
}

// RemoveByValues 删除数组中对应的多个目标
func (my *AnyArray[T]) RemoveByValues(targets ...T) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.removeByValues(targets...)
}

func (my *AnyArray[T]) every(fn func(item T) T) *AnyArray[T] {
	for idx := range my.data {
		v := fn(my.data[idx])
		my.data[idx] = v
	}

	return my
}

// Every 循环处理每一个
func (my *AnyArray[T]) Every(fn func(item T) T) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.every(fn)
}

func (my *AnyArray[T]) each(fn func(idx int, item T)) *AnyArray[T] {
	for idx := range my.data {
		fn(idx, my.data[idx])
	}

	return my
}

// Each 遍历数组
func (my *AnyArray[T]) Each(fn func(idx int, item T)) *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.each(fn)
}

func (my *AnyArray[T]) clean() *AnyArray[T] {
	my.data = make([]T, 0)

	return my
}

// Clean 清理数据
func (my *AnyArray[T]) Clean() *AnyArray[T] {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.clean()
}

func (my *AnyArray[T]) marshalJson() ([]byte, error) {
	return json.Marshal(&my.data)
}

// MarshalJSON 实现接口：json序列化
func (my *AnyArray[T]) MarshalJSON() ([]byte, error) {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.marshalJson()
}

func (my *AnyArray[T]) unmarshalJson(data []byte) error {
	return json.Unmarshal(data, &my.data)
}

// UnmarshalJSON 实现接口：json反序列化
func (my *AnyArray[T]) UnmarshalJSON(data []byte) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	return my.unmarshalJson(data)
}

// ToString 导出string
func (my *AnyArray[T]) ToString(formats ...string) string {
	var format = "%v"
	if len(formats) > 0 {
		format = formats[0]
	}

	return fmt.Sprintf(format, my.data)
}

// Cast 转换值类型
func Cast[SRC, DST any](aa *AnyArray[SRC], fn func(value SRC) DST) *AnyArray[DST] {
	if aa == nil {
		return nil
	}

	aa.mu.Lock()
	defer aa.mu.Unlock()

	data := make([]DST, len(aa.data))
	for i, v := range aa.data {
		data[i] = fn(v)
	}

	return New(data)
}

// ToAny converts any slice to []any
func ToAny(slice any) []any {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil
	}

	result := make([]any, v.Len())
	for i := range v.Len() {
		result[i] = v.Index(i).Interface()
	}

	return result
}
