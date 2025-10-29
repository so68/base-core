package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/so68/core/config"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data   map[string]*cacheItem
	mutex  sync.RWMutex
	config *config.CacheConfig
	logger *slog.Logger
}

// cacheItem 缓存项
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache(cfg *config.CacheConfig, logger *slog.Logger) (*MemoryCache, error) {
	cache := &MemoryCache{
		data:   make(map[string]*cacheItem),
		config: cfg,
		logger: logger,
	}

	// 启动清理协程
	go cache.cleanup()

	logger.Info("Memory cache connected successfully",
		slog.Int("max_memory", int(cfg.MaxMemory)),
		slog.Duration("cleanup_interval", cfg.CleanupInterval),
	)

	return cache, nil
}

// cleanup 定期清理过期项
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		for key, item := range m.data {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(m.data, key)
			}
		}
		m.mutex.Unlock()
	}
}

// serialize 序列化值
func (m *MemoryCache) serialize(value interface{}) interface{} {
	return value
}

// Get 获取值
func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key not found: %s", key)
	}

	switch v := item.value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// Set 设置值
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item := &cacheItem{
		value: m.serialize(value),
	}

	if expiration > 0 {
		item.expiration = time.Now().Add(expiration)
	}

	m.data[key] = item
	return nil
}

// Delete 删除键
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)
	return nil
}

// Exists 检查键是否存在
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return false, nil
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return false, nil
	}

	return true, nil
}

// MGet 批量获取
func (m *MemoryCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	values := make([]interface{}, len(keys))
	for i, key := range keys {
		item, exists := m.data[key]
		if !exists {
			values[i] = nil
			continue
		}

		if !item.expiration.IsZero() && time.Now().After(item.expiration) {
			delete(m.data, key)
			values[i] = nil
			continue
		}

		values[i] = item.value
	}

	return values, nil
}

// MSet 批量设置
func (m *MemoryCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for key, value := range pairs {
		item := &cacheItem{
			value: m.serialize(value),
		}

		if expiration > 0 {
			item.expiration = now.Add(expiration)
		}

		m.data[key] = item
	}

	return nil
}

// MDelete 批量删除
func (m *MemoryCache) MDelete(ctx context.Context, keys ...string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, key := range keys {
		delete(m.data, key)
	}

	return nil
}

// Increment 递增
func (m *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		m.data[key] = &cacheItem{value: delta}
		return delta, nil
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		m.data[key] = &cacheItem{value: delta}
		return delta, nil
	}

	var current int64
	switch v := item.value.(type) {
	case int:
		current = int64(v)
	case int64:
		current = v
	case float64:
		current = int64(v)
	default:
		return 0, fmt.Errorf("cannot increment non-numeric value")
	}

	newValue := current + delta
	m.data[key] = &cacheItem{value: newValue}
	return newValue, nil
}

// Decrement 递减
func (m *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return m.Increment(ctx, key, -delta)
}

// Expire 设置过期时间
func (m *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return fmt.Errorf("key not found: %s", key)
	}

	item.expiration = time.Now().Add(expiration)
	return nil
}

// TTL 获取剩余生存时间
func (m *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	if item.expiration.IsZero() {
		return -1, nil // 永不过期
	}

	if time.Now().After(item.expiration) {
		delete(m.data, key)
		return 0, fmt.Errorf("key not found: %s", key)
	}

	return time.Until(item.expiration), nil
}

// HGet 获取哈希字段值
func (m *MemoryCache) HGet(ctx context.Context, key, field string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key not found: %s", key)
	}

	hashMap, ok := item.value.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("key is not a hash: %s", key)
	}

	value, exists := hashMap[field]
	if !exists {
		return "", fmt.Errorf("field not found: %s.%s", key, field)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// HSet 设置哈希字段
func (m *MemoryCache) HSet(ctx context.Context, key string, pairs map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		item = &cacheItem{value: make(map[string]interface{})}
		m.data[key] = item
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		item = &cacheItem{value: make(map[string]interface{})}
		m.data[key] = item
	}

	hashMap, ok := item.value.(map[string]interface{})
	if !ok {
		hashMap = make(map[string]interface{})
		item.value = hashMap
	}

	for field, value := range pairs {
		hashMap[field] = m.serialize(value)
	}

	return nil
}

// HGetAll 获取所有哈希字段
func (m *MemoryCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return nil, fmt.Errorf("key not found: %s", key)
	}

	hashMap, ok := item.value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key is not a hash: %s", key)
	}

	result := make(map[string]string)
	for field, value := range hashMap {
		switch v := value.(type) {
		case string:
			result[field] = v
		default:
			result[field] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}

// HDelete 删除哈希字段
func (m *MemoryCache) HDelete(ctx context.Context, key string, fields ...string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return fmt.Errorf("key not found: %s", key)
	}

	hashMap, ok := item.value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("key is not a hash: %s", key)
	}

	for _, field := range fields {
		delete(hashMap, field)
	}

	return nil
}

// LPush 左推入列表
func (m *MemoryCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		item = &cacheItem{value: make([]interface{}, 0)}
		m.data[key] = item
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		item = &cacheItem{value: make([]interface{}, 0)}
		m.data[key] = item
	}

	list, ok := item.value.([]interface{})
	if !ok {
		list = make([]interface{}, 0)
		item.value = list
	}

	// 左推入
	for _, value := range values {
		list = append([]interface{}{m.serialize(value)}, list...)
	}

	item.value = list
	return nil
}

// RPush 右推入列表
func (m *MemoryCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		item = &cacheItem{value: make([]interface{}, 0)}
		m.data[key] = item
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		item = &cacheItem{value: make([]interface{}, 0)}
		m.data[key] = item
	}

	list, ok := item.value.([]interface{})
	if !ok {
		list = make([]interface{}, 0)
		item.value = list
	}

	// 右推入
	for _, value := range values {
		list = append(list, m.serialize(value))
	}

	item.value = list
	return nil
}

// LPop 左弹出列表
func (m *MemoryCache) LPop(ctx context.Context, key string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key not found: %s", key)
	}

	list, ok := item.value.([]interface{})
	if !ok || len(list) == 0 {
		return "", fmt.Errorf("list is empty: %s", key)
	}

	value := list[0]
	list = list[1:]
	item.value = list

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// RPop 右弹出列表
func (m *MemoryCache) RPop(ctx context.Context, key string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key not found: %s", key)
	}

	list, ok := item.value.([]interface{})
	if !ok || len(list) == 0 {
		return "", fmt.Errorf("list is empty: %s", key)
	}

	value := list[len(list)-1]
	list = list[:len(list)-1]
	item.value = list

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// LRange 获取列表范围
func (m *MemoryCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return nil, fmt.Errorf("key not found: %s", key)
	}

	list, ok := item.value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("key is not a list: %s", key)
	}

	length := int64(len(list))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}, nil
	}

	result := make([]string, stop-start+1)
	for i := start; i <= stop; i++ {
		switch v := list[i].(type) {
		case string:
			result[i-start] = v
		default:
			result[i-start] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}

// SAdd 添加集合成员
func (m *MemoryCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		item = &cacheItem{value: make(map[interface{}]bool)}
		m.data[key] = item
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		item = &cacheItem{value: make(map[interface{}]bool)}
		m.data[key] = item
	}

	set, ok := item.value.(map[interface{}]bool)
	if !ok {
		set = make(map[interface{}]bool)
		item.value = set
	}

	for _, member := range members {
		set[m.serialize(member)] = true
	}

	return nil
}

// SRem 删除集合成员
func (m *MemoryCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return fmt.Errorf("key not found: %s", key)
	}

	set, ok := item.value.(map[interface{}]bool)
	if !ok {
		return fmt.Errorf("key is not a set: %s", key)
	}

	for _, member := range members {
		delete(set, m.serialize(member))
	}

	return nil
}

// SMembers 获取集合所有成员
func (m *MemoryCache) SMembers(ctx context.Context, key string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return nil, fmt.Errorf("key not found: %s", key)
	}

	set, ok := item.value.(map[interface{}]bool)
	if !ok {
		return nil, fmt.Errorf("key is not a set: %s", key)
	}

	result := make([]string, 0, len(set))
	for member := range set {
		switch v := member.(type) {
		case string:
			result = append(result, v)
		default:
			result = append(result, fmt.Sprintf("%v", v))
		}
	}

	return result, nil
}

// SIsMember 检查集合成员
func (m *MemoryCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return false, fmt.Errorf("key not found: %s", key)
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return false, fmt.Errorf("key not found: %s", key)
	}

	set, ok := item.value.(map[interface{}]bool)
	if !ok {
		return false, fmt.Errorf("key is not a set: %s", key)
	}

	return set[m.serialize(member)], nil
}

// HealthCheck 健康检查
func (m *MemoryCache) HealthCheck(ctx context.Context) error {
	// 内存缓存总是健康的
	m.logger.Info("Memory cache healthy")
	return nil
}

// Close 关闭连接
func (m *MemoryCache) Close() error {
	// 内存缓存不需要关闭连接
	return nil
}
