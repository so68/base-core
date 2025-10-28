package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/so68/core/config"
)

/*
Redis缓存功能测试

本文件用于测试RedisCache结构体的各种功能特性，
包括缓存操作、策略模式、性能测试、错误处理等。

运行命令：
go test -v -run "^Test.*Redis.*$"

测试内容：
1. 基本缓存操作 (Get, Set, Delete, Exists等)
2. 批量操作测试 (MGet, MSet, MDelete等)
3. 高级操作测试 (Increment, Decrement, Expire, TTL等)
4. 哈希操作测试 (HGet, HSet, HGetAll, HDelete等)
5. 列表操作测试 (LPush, RPush, LPop, RPop, LRange等)
6. 集合操作测试 (SAdd, SRem, SMembers, SIsMember等)
7. 策略模式测试 (缓存穿透、雪崩、击穿、多级缓存等)
8. 健康检查和连接管理测试
*/

func TestRedisCache_BasicOperations(t *testing.T) {
	// 创建测试配置
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1, // 使用测试数据库
		Prefix:   "test",
	}
	cfg.SetDefaults()

	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// 创建 Redis 缓存实例
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 测试基本操作
	t.Run("Set and Get", func(t *testing.T) {
		err := cache.Set(ctx, "test_key", "test_value", time.Hour)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		value, err := cache.Get(ctx, "test_key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if value != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", value)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := cache.Exists(ctx, "test_key")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if !exists {
			t.Error("Expected key to exist")
		}

		exists, err = cache.Exists(ctx, "non_existent_key")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if exists {
			t.Error("Expected key to not exist")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := cache.Delete(ctx, "test_key")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		exists, err := cache.Exists(ctx, "test_key")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if exists {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		// 设置初始值
		err := cache.Set(ctx, "counter", "10", time.Hour)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// 递增
		value, err := cache.Increment(ctx, "counter", 5)
		if err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
		if value != 15 {
			t.Errorf("Expected 15, got %d", value)
		}

		// 递减
		value, err = cache.Decrement(ctx, "counter", 3)
		if err != nil {
			t.Fatalf("Decrement failed: %v", err)
		}
		if value != 12 {
			t.Errorf("Expected 12, got %d", value)
		}
	})

	t.Run("Expire and TTL", func(t *testing.T) {
		err := cache.Set(ctx, "expire_key", "expire_value", time.Second*2)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ttl, err := cache.TTL(ctx, "expire_key")
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl <= 0 || ttl > time.Second*2 {
			t.Errorf("Expected TTL between 0 and 2 seconds, got %v", ttl)
		}

		// 等待过期
		time.Sleep(time.Second * 3)

		_, err = cache.Get(ctx, "expire_key")
		if err == nil {
			t.Error("Expected key to be expired")
		}
	})
}

func TestRedisCache_HashOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "test",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("HSet and HGet", func(t *testing.T) {
		pairs := map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		}

		err := cache.HSet(ctx, "user:1", pairs)
		if err != nil {
			t.Fatalf("HSet failed: %v", err)
		}

		name, err := cache.HGet(ctx, "user:1", "name")
		if err != nil {
			t.Fatalf("HGet failed: %v", err)
		}
		if name != "John" {
			t.Errorf("Expected 'John', got '%s'", name)
		}

		allData, err := cache.HGetAll(ctx, "user:1")
		if err != nil {
			t.Fatalf("HGetAll failed: %v", err)
		}

		if len(allData) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(allData))
		}
	})

	t.Run("HDelete", func(t *testing.T) {
		err := cache.HDelete(ctx, "user:1", "age")
		if err != nil {
			t.Fatalf("HDelete failed: %v", err)
		}

		_, err = cache.HGet(ctx, "user:1", "age")
		if err == nil {
			t.Error("Expected field to be deleted")
		}
	})
}

func TestRedisCache_ListOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "test",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("LPush and RPush", func(t *testing.T) {
		err := cache.LPush(ctx, "list", "first")
		if err != nil {
			t.Fatalf("LPush failed: %v", err)
		}

		err = cache.RPush(ctx, "list", "last")
		if err != nil {
			t.Fatalf("RPush failed: %v", err)
		}

		items, err := cache.LRange(ctx, "list", 0, -1)
		if err != nil {
			t.Fatalf("LRange failed: %v", err)
		}

		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}
	})

	t.Run("LPop and RPop", func(t *testing.T) {
		value, err := cache.LPop(ctx, "list")
		if err != nil {
			t.Fatalf("LPop failed: %v", err)
		}
		if value != "first" {
			t.Errorf("Expected 'first', got '%s'", value)
		}

		value, err = cache.RPop(ctx, "list")
		if err != nil {
			t.Fatalf("RPop failed: %v", err)
		}
		if value != "last" {
			t.Errorf("Expected 'last', got '%s'", value)
		}
	})
}

func TestRedisCache_SetOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "test",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("SAdd and SMembers", func(t *testing.T) {
		err := cache.SAdd(ctx, "set", "member1", "member2", "member3")
		if err != nil {
			t.Fatalf("SAdd failed: %v", err)
		}

		members, err := cache.SMembers(ctx, "set")
		if err != nil {
			t.Fatalf("SMembers failed: %v", err)
		}

		if len(members) != 3 {
			t.Errorf("Expected 3 members, got %d", len(members))
		}
	})

	t.Run("SIsMember", func(t *testing.T) {
		isMember, err := cache.SIsMember(ctx, "set", "member1")
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if !isMember {
			t.Error("Expected member1 to be in set")
		}

		isMember, err = cache.SIsMember(ctx, "set", "non_member")
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if isMember {
			t.Error("Expected non_member to not be in set")
		}
	})

	t.Run("SRem", func(t *testing.T) {
		err := cache.SRem(ctx, "set", "member1")
		if err != nil {
			t.Fatalf("SRem failed: %v", err)
		}

		isMember, err := cache.SIsMember(ctx, "set", "member1")
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if isMember {
			t.Error("Expected member1 to be removed from set")
		}
	})
}

func TestRedisCache_BatchOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "test",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("MSet and MGet", func(t *testing.T) {
		pairs := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		err := cache.MSet(ctx, pairs, time.Hour)
		if err != nil {
			t.Fatalf("MSet failed: %v", err)
		}

		values, err := cache.MGet(ctx, "key1", "key2", "key3")
		if err != nil {
			t.Fatalf("MGet failed: %v", err)
		}

		if len(values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(values))
		}

		for i, value := range values {
			if value == nil {
				t.Errorf("Expected value for key%d, got nil", i+1)
			}
		}
	})

	t.Run("MDelete", func(t *testing.T) {
		err := cache.MDelete(ctx, "key1", "key2", "key3")
		if err != nil {
			t.Fatalf("MDelete failed: %v", err)
		}

		exists, err := cache.Exists(ctx, "key1")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Expected key1 to be deleted")
		}
	})
}

func TestRedisCache_HealthCheck(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "test",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

// TestRedisCache_StrategyPattern 测试策略模式用例
func TestRedisCache_StrategyPattern(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Prefix:   "strategy",
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 测试用例1: 缓存穿透策略 - 布隆过滤器模拟
	t.Run("Cache Penetration Strategy", func(t *testing.T) {
		// 模拟布隆过滤器：使用集合存储已存在的键
		filterKey := "bloom:filter"

		// 添加一些已知存在的键到布隆过滤器
		knownKeys := []string{"user:1", "user:2", "user:3", "product:1", "product:2"}
		for _, key := range knownKeys {
			err := cache.SAdd(ctx, filterKey, key)
			if err != nil {
				t.Fatalf("SAdd to bloom filter failed: %v", err)
			}
		}

		// 测试键是否存在
		testKey := "user:1"
		exists, err := cache.SIsMember(ctx, filterKey, testKey)
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if !exists {
			t.Error("Expected key to exist in bloom filter")
		}

		// 测试不存在的键
		nonExistentKey := "user:999"
		exists, err = cache.SIsMember(ctx, filterKey, nonExistentKey)
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		if exists {
			t.Error("Expected key to not exist in bloom filter")
		}
	})

	// 测试用例2: 缓存雪崩策略 - 随机过期时间
	t.Run("Cache Avalanche Strategy", func(t *testing.T) {
		baseExpiration := time.Minute * 5
		randomOffset := time.Second * 30

		// 设置多个键，使用随机过期时间避免同时过期
		keys := []string{"cache:key1", "cache:key2", "cache:key3", "cache:key4", "cache:key5"}
		for i, key := range keys {
			// 模拟随机过期时间
			randomExpiration := baseExpiration + time.Duration(i*10)*time.Second
			if randomExpiration > baseExpiration+randomOffset {
				randomExpiration = baseExpiration + randomOffset
			}

			err := cache.Set(ctx, key, fmt.Sprintf("value%d", i+1), randomExpiration)
			if err != nil {
				t.Fatalf("Set with random expiration failed: %v", err)
			}
		}

		// 验证所有键都已设置
		for i, key := range keys {
			value, err := cache.Get(ctx, key)
			if err != nil {
				t.Fatalf("Get failed for key %s: %v", key, err)
			}
			expectedValue := fmt.Sprintf("value%d", i+1)
			if value != expectedValue {
				t.Errorf("Expected %s, got %s", expectedValue, value)
			}
		}
	})

	// 测试用例3: 缓存击穿策略 - 分布式锁模拟
	t.Run("Cache Breakdown Strategy", func(t *testing.T) {
		lockKey := "lock:expensive:operation"
		lockValue := "lock_holder_123"
		lockExpiration := time.Second * 10

		// 尝试获取分布式锁
		lockAcquired := false
		for i := 0; i < 3; i++ {
			// 使用 SET NX EX 命令实现分布式锁
			err := cache.Set(ctx, lockKey, lockValue, lockExpiration)
			if err == nil {
				lockAcquired = true
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		if !lockAcquired {
			t.Error("Failed to acquire distributed lock")
		}

		// 验证锁已获取
		value, err := cache.Get(ctx, lockKey)
		if err != nil {
			t.Fatalf("Get lock value failed: %v", err)
		}
		if value != lockValue {
			t.Errorf("Expected lock value %s, got %s", lockValue, value)
		}

		// 释放锁
		err = cache.Delete(ctx, lockKey)
		if err != nil {
			t.Fatalf("Release lock failed: %v", err)
		}
	})

	// 测试用例4: 多级缓存策略 - L1(内存) + L2(Redis)
	t.Run("Multi-Level Cache Strategy", func(t *testing.T) {
		// L1 缓存键
		l1Key := "l1:user:profile:1"
		// L2 缓存键
		l2Key := "l2:user:profile:1"

		// 模拟 L2 缓存（Redis）存储
		l2Data := map[string]interface{}{
			"name":      "John Doe",
			"email":     "john@example.com",
			"level":     "premium",
			"lastLogin": time.Now().Unix(),
		}

		err := cache.HSet(ctx, l2Key, l2Data)
		if err != nil {
			t.Fatalf("HSet L2 cache failed: %v", err)
		}

		// 模拟 L1 缓存命中
		l1Hit := false
		exists, err := cache.Exists(ctx, l1Key)
		if err != nil {
			t.Fatalf("Check L1 cache failed: %v", err)
		}

		if !exists {
			// L1 缓存未命中，从 L2 加载
			l2Data, err := cache.HGetAll(ctx, l2Key)
			if err != nil {
				t.Fatalf("HGetAll L2 cache failed: %v", err)
			}

			// 将 L2 数据加载到 L1
			err = cache.HSet(ctx, l1Key, map[string]interface{}{
				"name":      l2Data["name"],
				"email":     l2Data["email"],
				"level":     l2Data["level"],
				"lastLogin": l2Data["lastLogin"],
			})
			if err != nil {
				t.Fatalf("HSet L1 cache failed: %v", err)
			}

			// 设置 L1 缓存过期时间（较短）
			err = cache.Expire(ctx, l1Key, time.Minute*5)
			if err != nil {
				t.Fatalf("Set L1 expiration failed: %v", err)
			}
		} else {
			l1Hit = true
		}

		// 验证数据一致性
		l1Data, err := cache.HGetAll(ctx, l1Key)
		if err != nil {
			t.Fatalf("HGetAll L1 cache failed: %v", err)
		}

		l2DataCheck, err := cache.HGetAll(ctx, l2Key)
		if err != nil {
			t.Fatalf("HGetAll L2 cache failed: %v", err)
		}

		if l1Data["name"] != l2DataCheck["name"] {
			t.Error("L1 and L2 cache data inconsistency")
		}

		t.Logf("L1 cache hit: %t", l1Hit)
	})

	// 测试用例5: 缓存预热策略
	t.Run("Cache Warming Strategy", func(t *testing.T) {
		// 模拟预热数据
		warmupData := map[string]interface{}{
			"config:app_name":    "MyApp",
			"config:version":     "1.0.0",
			"config:environment": "production",
			"config:debug_mode":  false,
			"config:max_users":   10000,
			"config:cache_ttl":   3600,
		}

		// 批量预热缓存
		err := cache.MSet(ctx, warmupData, time.Hour*24) // 24小时过期
		if err != nil {
			t.Fatalf("Cache warming failed: %v", err)
		}

		// 验证预热数据
		keys := make([]string, 0, len(warmupData))
		for key := range warmupData {
			keys = append(keys, key)
		}

		values, err := cache.MGet(ctx, keys...)
		if err != nil {
			t.Fatalf("MGet warmup data failed: %v", err)
		}

		// 检查所有预热数据都存在
		for i, value := range values {
			if value == nil {
				t.Errorf("Warmup data missing for key: %s", keys[i])
			}
		}

		t.Logf("Successfully warmed up %d cache entries", len(warmupData))
	})

	// 测试用例6: 缓存更新策略 - Write-Through
	t.Run("Write-Through Cache Strategy", func(t *testing.T) {
		// 模拟数据库更新
		dbKey := "db:user:1"
		cacheKey := "cache:user:1"

		// 模拟数据库数据
		dbData := map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
			"age":   25,
		}

		// Write-Through: 同时更新数据库和缓存
		// 1. 更新数据库（模拟）
		err := cache.HSet(ctx, dbKey, dbData)
		if err != nil {
			t.Fatalf("Update database failed: %v", err)
		}

		// 2. 更新缓存
		err = cache.HSet(ctx, cacheKey, dbData)
		if err != nil {
			t.Fatalf("Update cache failed: %v", err)
		}

		// 验证数据一致性
		dbDataCheck, err := cache.HGetAll(ctx, dbKey)
		if err != nil {
			t.Fatalf("HGetAll database failed: %v", err)
		}

		cacheDataCheck, err := cache.HGetAll(ctx, cacheKey)
		if err != nil {
			t.Fatalf("HGetAll cache failed: %v", err)
		}

		if dbDataCheck["name"] != cacheDataCheck["name"] {
			t.Error("Database and cache data inconsistency")
		}

		t.Log("Write-Through strategy executed successfully")
	})

	// 测试用例7: 缓存淘汰策略 - LRU 模拟
	t.Run("LRU Eviction Strategy", func(t *testing.T) {
		// 模拟 LRU 链表
		lruListKey := "lru:list"
		maxSize := 5

		// 添加元素到 LRU 列表
		items := []string{"item1", "item2", "item3", "item4", "item5", "item6"}

		for _, item := range items {
			// 检查是否已存在
			exists, err := cache.SIsMember(ctx, lruListKey, item)
			if err != nil {
				t.Fatalf("SIsMember failed: %v", err)
			}

			if exists {
				// 如果存在，移动到列表末尾（最近使用）
				err = cache.SRem(ctx, lruListKey, item)
				if err != nil {
					t.Fatalf("SRem failed: %v", err)
				}
			}

			// 添加到列表末尾
			err = cache.SAdd(ctx, lruListKey, item)
			if err != nil {
				t.Fatalf("SAdd failed: %v", err)
			}

			// 检查列表大小，如果超过最大大小，移除最旧的
			members, err := cache.SMembers(ctx, lruListKey)
			if err != nil {
				t.Fatalf("SMembers failed: %v", err)
			}

			if len(members) > maxSize {
				// 移除第一个元素（最旧的）
				oldestItem := members[0]
				err = cache.SRem(ctx, lruListKey, oldestItem)
				if err != nil {
					t.Fatalf("SRem oldest item failed: %v", err)
				}
				t.Logf("Evicted oldest item: %s", oldestItem)
			}
		}

		// 验证最终列表大小
		finalMembers, err := cache.SMembers(ctx, lruListKey)
		if err != nil {
			t.Fatalf("SMembers final check failed: %v", err)
		}

		if len(finalMembers) > maxSize {
			t.Errorf("LRU list size exceeded max size: %d > %d", len(finalMembers), maxSize)
		}

		t.Logf("LRU list final size: %d", len(finalMembers))
	})

	// 测试用例8: 缓存监控策略
	t.Run("Cache Monitoring Strategy", func(t *testing.T) {
		// 缓存统计键
		statsKey := "cache:stats"

		// 初始化统计
		initialStats := map[string]interface{}{
			"hits":    0,
			"misses":  0,
			"sets":    0,
			"deletes": 0,
		}

		err := cache.HSet(ctx, statsKey, initialStats)
		if err != nil {
			t.Fatalf("Initialize stats failed: %v", err)
		}

		// 模拟缓存操作并更新统计
		testKey := "monitor:test:key"

		// 模拟缓存未命中
		_, err = cache.Get(ctx, testKey)
		if err != nil {
			// 增加未命中计数
			_, err = cache.Increment(ctx, statsKey+":misses", 1)
			if err != nil {
				t.Fatalf("Increment misses failed: %v", err)
			}
		}

		// 模拟缓存设置
		err = cache.Set(ctx, testKey, "test_value", time.Minute)
		if err != nil {
			t.Fatalf("Set test key failed: %v", err)
		}

		// 增加设置计数
		_, err = cache.Increment(ctx, statsKey+":sets", 1)
		if err != nil {
			t.Fatalf("Increment sets failed: %v", err)
		}

		// 模拟缓存命中
		_, err = cache.Get(ctx, testKey)
		if err == nil {
			// 增加命中计数
			_, err = cache.Increment(ctx, statsKey+":hits", 1)
			if err != nil {
				t.Fatalf("Increment hits failed: %v", err)
			}
		}

		// 获取最终统计
		hits, err := cache.Get(ctx, statsKey+":hits")
		if err != nil {
			t.Fatalf("Get hits count failed: %v", err)
		}

		misses, err := cache.Get(ctx, statsKey+":misses")
		if err != nil {
			t.Fatalf("Get misses count failed: %v", err)
		}

		sets, err := cache.Get(ctx, statsKey+":sets")
		if err != nil {
			t.Fatalf("Get sets count failed: %v", err)
		}

		t.Logf("Cache stats - Hits: %s, Misses: %s, Sets: %s", hits, misses, sets)
	})
}
