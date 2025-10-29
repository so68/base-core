package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/so68/core/config"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redis.Client
	config *config.CacheConfig
	logger *slog.Logger
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(cfg *config.CacheConfig, logger *slog.Logger) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.Database,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Redis cache connected successfully",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", strconv.Itoa(cfg.Database)),
	)

	return &RedisCache{
		client: rdb,
		config: cfg,
		logger: logger,
	}, nil
}

// getKey 获取带前缀的键
func (r *RedisCache) getKey(key string) string {
	if r.config.Prefix != "" {
		return r.config.Prefix + ":" + key
	}
	return key
}

// serialize 序列化值
func (r *RedisCache) serialize(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("failed to serialize value: %w", err)
		}
		return string(data), nil
	}
}

// deserialize 反序列化值
func (r *RedisCache) deserialize(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}

// Get 获取值
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result := r.client.Get(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return result.Val(), nil
}

// Set 设置值
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	serialized, err := r.serialize(value)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, r.getKey(key), serialized, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete 删除键
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, r.getKey(key)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return result.Val() > 0, nil
}

// MGet 批量获取
func (r *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.getKey(key)
	}

	result := r.client.MGet(ctx, redisKeys...)
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to mget keys: %w", err)
	}

	values := make([]interface{}, len(result.Val()))
	for i, val := range result.Val() {
		if val == nil {
			values[i] = nil
		} else {
			values[i] = val
		}
	}

	return values, nil
}

// MSet 批量设置
func (r *RedisCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range pairs {
		serialized, err := r.serialize(value)
		if err != nil {
			return err
		}

		if expiration > 0 {
			pipe.Set(ctx, r.getKey(key), serialized, expiration)
		} else {
			pipe.Set(ctx, r.getKey(key), serialized, 0)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mset keys: %w", err)
	}

	return nil
}

// MDelete 批量删除
func (r *RedisCache) MDelete(ctx context.Context, keys ...string) error {
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.getKey(key)
	}

	err := r.client.Del(ctx, redisKeys...).Err()
	if err != nil {
		return fmt.Errorf("failed to mdelete keys: %w", err)
	}
	return nil
}

// Increment 递增
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	result := r.client.IncrBy(ctx, r.getKey(key), delta)
	if err := result.Err(); err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result.Val(), nil
}

// Decrement 递减
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	result := r.client.DecrBy(ctx, r.getKey(key), delta)
	if err := result.Err(); err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return result.Val(), nil
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, r.getKey(key), expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// TTL 获取剩余生存时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	result := r.client.TTL(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		return 0, fmt.Errorf("failed to get ttl for key %s: %w", key, err)
	}
	return result.Val(), nil
}

// HGet 获取哈希字段值
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	result := r.client.HGet(ctx, r.getKey(key), field)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("field not found: %s.%s", key, field)
		}
		return "", fmt.Errorf("failed to hget field %s.%s: %w", key, field, err)
	}
	return result.Val(), nil
}

// HSet 设置哈希字段
func (r *RedisCache) HSet(ctx context.Context, key string, pairs map[string]interface{}) error {
	values := make(map[string]interface{})
	for field, value := range pairs {
		serialized, err := r.serialize(value)
		if err != nil {
			return err
		}
		values[field] = serialized
	}

	err := r.client.HSet(ctx, r.getKey(key), values).Err()
	if err != nil {
		return fmt.Errorf("failed to hset key %s: %w", key, err)
	}
	return nil
}

// HGetAll 获取所有哈希字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result := r.client.HGetAll(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to hgetall key %s: %w", key, err)
	}
	return result.Val(), nil
}

// HDelete 删除哈希字段
func (r *RedisCache) HDelete(ctx context.Context, key string, fields ...string) error {
	err := r.client.HDel(ctx, r.getKey(key), fields...).Err()
	if err != nil {
		return fmt.Errorf("failed to hdel fields from key %s: %w", key, err)
	}
	return nil
}

// LPush 左推入列表
func (r *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	serializedValues := make([]interface{}, len(values))
	for i, value := range values {
		serialized, err := r.serialize(value)
		if err != nil {
			return err
		}
		serializedValues[i] = serialized
	}

	err := r.client.LPush(ctx, r.getKey(key), serializedValues...).Err()
	if err != nil {
		return fmt.Errorf("failed to lpush to key %s: %w", key, err)
	}
	return nil
}

// RPush 右推入列表
func (r *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	serializedValues := make([]interface{}, len(values))
	for i, value := range values {
		serialized, err := r.serialize(value)
		if err != nil {
			return err
		}
		serializedValues[i] = serialized
	}

	err := r.client.RPush(ctx, r.getKey(key), serializedValues...).Err()
	if err != nil {
		return fmt.Errorf("failed to rpush to key %s: %w", key, err)
	}
	return nil
}

// LPop 左弹出列表
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	result := r.client.LPop(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list is empty: %s", key)
		}
		return "", fmt.Errorf("failed to lpop from key %s: %w", key, err)
	}
	return result.Val(), nil
}

// RPop 右弹出列表
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	result := r.client.RPop(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list is empty: %s", key)
		}
		return "", fmt.Errorf("failed to rpop from key %s: %w", key, err)
	}
	return result.Val(), nil
}

// LRange 获取列表范围
func (r *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result := r.client.LRange(ctx, r.getKey(key), start, stop)
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to lrange key %s: %w", key, err)
	}
	return result.Val(), nil
}

// SAdd 添加集合成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	serializedMembers := make([]interface{}, len(members))
	for i, member := range members {
		serialized, err := r.serialize(member)
		if err != nil {
			return err
		}
		serializedMembers[i] = serialized
	}

	err := r.client.SAdd(ctx, r.getKey(key), serializedMembers...).Err()
	if err != nil {
		return fmt.Errorf("failed to sadd to key %s: %w", key, err)
	}
	return nil
}

// SRem 删除集合成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	serializedMembers := make([]interface{}, len(members))
	for i, member := range members {
		serialized, err := r.serialize(member)
		if err != nil {
			return err
		}
		serializedMembers[i] = serialized
	}

	err := r.client.SRem(ctx, r.getKey(key), serializedMembers...).Err()
	if err != nil {
		return fmt.Errorf("failed to srem from key %s: %w", key, err)
	}
	return nil
}

// SMembers 获取集合所有成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	result := r.client.SMembers(ctx, r.getKey(key))
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to smembers key %s: %w", key, err)
	}
	return result.Val(), nil
}

// SIsMember 检查集合成员
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	serialized, err := r.serialize(member)
	if err != nil {
		return false, err
	}

	result := r.client.SIsMember(ctx, r.getKey(key), serialized)
	if err := result.Err(); err != nil {
		return false, fmt.Errorf("failed to sismember key %s: %w", key, err)
	}
	return result.Val(), nil
}

// HealthCheck 健康检查
func (r *RedisCache) HealthCheck(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// 检查连接池状态
	stats := r.client.PoolStats()
	r.logger.Info("Redis connection pool stats",
		slog.Int("total_conns", int(stats.TotalConns)),
		slog.Int("idle_conns", int(stats.IdleConns)),
	)

	return nil
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}
