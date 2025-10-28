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
内存缓存功能测试

本文件用于测试MemoryCache结构体的各种功能特性，
包括缓存操作、过期处理、内存管理、并发安全等。

运行命令：
go test -v -run "^Test.*Memory.*$"

测试内容：
1. 基本缓存操作 (Get, Set, Delete, Exists等)
2. 批量操作测试 (MGet, MSet, MDelete等)
3. 高级操作测试 (Increment, Decrement, Expire, TTL等)
4. 哈希操作测试 (HGet, HSet, HGetAll, HDelete等)
5. 列表操作测试 (LPush, RPush, LPop, RPop, LRange等)
6. 集合操作测试 (SAdd, SRem, SMembers, SIsMember等)
7. 过期处理测试 (自动清理、TTL计算等)
8. 并发安全测试 (多协程访问等)
9. 内存管理测试 (内存限制、清理机制等)
10. 健康检查和错误处理测试
*/

func TestMemoryCache_BasicOperations(t *testing.T) {
	// 创建测试配置
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024, // 100MB
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// 创建内存缓存实例
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
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
		// 设置初始值（使用数字类型）
		err := cache.Set(ctx, "counter", 10, time.Hour)
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

func TestMemoryCache_HashOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
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

func TestMemoryCache_ListOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
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

		if items[0] != "first" {
			t.Errorf("Expected first item to be 'first', got '%s'", items[0])
		}
		if items[1] != "last" {
			t.Errorf("Expected second item to be 'last', got '%s'", items[1])
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

	t.Run("LRange with bounds", func(t *testing.T) {
		// 添加多个元素
		err := cache.RPush(ctx, "range_list", "a", "b", "c", "d", "e")
		if err != nil {
			t.Fatalf("RPush failed: %v", err)
		}

		// 测试范围查询
		items, err := cache.LRange(ctx, "range_list", 1, 3)
		if err != nil {
			t.Fatalf("LRange failed: %v", err)
		}

		expected := []string{"b", "c", "d"}
		if len(items) != len(expected) {
			t.Errorf("Expected %d items, got %d", len(expected), len(items))
		}

		for i, item := range items {
			if item != expected[i] {
				t.Errorf("Expected item %d to be '%s', got '%s'", i, expected[i], item)
			}
		}
	})
}

func TestMemoryCache_SetOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
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

		// 检查成员是否存在
		expectedMembers := map[string]bool{
			"member1": true,
			"member2": true,
			"member3": true,
		}

		for _, member := range members {
			if !expectedMembers[member] {
				t.Errorf("Unexpected member: %s", member)
			}
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

func TestMemoryCache_BatchOperations(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
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

func TestMemoryCache_ExpirationHandling(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Second, // 短清理间隔用于测试
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Automatic Expiration", func(t *testing.T) {
		// 设置短期过期的键
		err := cache.Set(ctx, "short_lived", "value", time.Millisecond*100)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// 立即检查应该存在
		exists, err := cache.Exists(ctx, "short_lived")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Expected key to exist immediately after setting")
		}

		// 等待过期
		time.Sleep(time.Millisecond * 200)

		// 检查是否已过期
		exists, err = cache.Exists(ctx, "short_lived")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Expected key to be expired")
		}
	})

	t.Run("TTL Calculation", func(t *testing.T) {
		// 设置带过期时间的键
		err := cache.Set(ctx, "ttl_test", "value", time.Second*5)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ttl, err := cache.TTL(ctx, "ttl_test")
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl <= 0 || ttl > time.Second*5 {
			t.Errorf("Expected TTL between 0 and 5 seconds, got %v", ttl)
		}

		// 等待一段时间后再次检查
		time.Sleep(time.Second * 2)

		ttl, err = cache.TTL(ctx, "ttl_test")
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl <= 0 || ttl > time.Second*3 {
			t.Errorf("Expected TTL between 0 and 3 seconds after 2s wait, got %v", ttl)
		}
	})

	t.Run("Permanent Keys", func(t *testing.T) {
		// 设置永不过期的键
		err := cache.Set(ctx, "permanent", "value", 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ttl, err := cache.TTL(ctx, "permanent")
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}

		if ttl != -1 {
			t.Errorf("Expected TTL to be -1 for permanent key, got %v", ttl)
		}
	})
}

func TestMemoryCache_Concurrency(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Concurrent Reads and Writes", func(t *testing.T) {
		const numGoroutines = 10
		const numOperations = 100

		// 启动多个协程进行并发读写
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("concurrent_key_%d_%d", goroutineID, j)
					value := fmt.Sprintf("value_%d_%d", goroutineID, j)

					// 写入
					err := cache.Set(ctx, key, value, time.Hour)
					if err != nil {
						t.Errorf("Set failed in goroutine %d: %v", goroutineID, err)
						return
					}

					// 读取
					retrievedValue, err := cache.Get(ctx, key)
					if err != nil {
						t.Errorf("Get failed in goroutine %d: %v", goroutineID, err)
						return
					}

					if retrievedValue != value {
						t.Errorf("Value mismatch in goroutine %d: expected %s, got %s", goroutineID, value, retrievedValue)
						return
					}
				}
			}(i)
		}

		// 等待所有协程完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("Concurrent Increment", func(t *testing.T) {
		const numGoroutines = 5
		const incrementPerGoroutine = 20

		// 初始化计数器
		err := cache.Set(ctx, "concurrent_counter", 0, time.Hour)
		if err != nil {
			t.Fatalf("Set counter failed: %v", err)
		}

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < incrementPerGoroutine; j++ {
					_, err := cache.Increment(ctx, "concurrent_counter", 1)
					if err != nil {
						t.Errorf("Increment failed: %v", err)
						return
					}
				}
			}()
		}

		// 等待所有协程完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// 检查最终值
		finalValue, err := cache.Get(ctx, "concurrent_counter")
		if err != nil {
			t.Fatalf("Get final value failed: %v", err)
		}

		expectedValue := numGoroutines * incrementPerGoroutine
		if finalValue != fmt.Sprintf("%d", expectedValue) {
			t.Errorf("Expected final value %d, got %s", expectedValue, finalValue)
		}
	})
}

func TestMemoryCache_MemoryManagement(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       1024, // 1KB 用于测试
		CleanupInterval: time.Millisecond * 100,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Memory Usage", func(t *testing.T) {
		// 添加一些数据
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("memory_test_%d", i)
			value := fmt.Sprintf("value_%d_%s", i, "x") // 创建一些数据
			err := cache.Set(ctx, key, value, time.Hour)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		// 验证数据存在
		exists, err := cache.Exists(ctx, "memory_test_0")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Expected data to exist")
		}
	})

	t.Run("Cleanup Mechanism", func(t *testing.T) {
		// 设置一些短期过期的数据
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("cleanup_test_%d", i)
			err := cache.Set(ctx, key, "value", time.Millisecond*50)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		// 等待过期
		time.Sleep(time.Millisecond * 100)

		// 等待清理
		time.Sleep(time.Millisecond * 200)

		// 检查是否被清理
		exists, err := cache.Exists(ctx, "cleanup_test_0")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Expected expired data to be cleaned up")
		}
	})
}

func TestMemoryCache_ErrorHandling(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Non-existent Key Operations", func(t *testing.T) {
		// 获取不存在的键
		_, err := cache.Get(ctx, "non_existent")
		if err == nil {
			t.Error("Expected error when getting non-existent key")
		}

		// 删除不存在的键（应该不报错）
		err = cache.Delete(ctx, "non_existent")
		if err != nil {
			t.Errorf("Delete non-existent key should not error: %v", err)
		}

		// 检查不存在的键
		exists, err := cache.Exists(ctx, "non_existent")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Expected non-existent key to not exist")
		}
	})

	t.Run("Invalid Data Type Operations", func(t *testing.T) {
		// 设置一个字符串值
		err := cache.Set(ctx, "string_key", "string_value", time.Hour)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// 尝试对字符串进行哈希操作
		_, err = cache.HGet(ctx, "string_key", "field")
		if err == nil {
			t.Error("Expected error when HGet on non-hash key")
		}

		// 尝试对字符串进行列表操作
		_, err = cache.LPop(ctx, "string_key")
		if err == nil {
			t.Error("Expected error when LPop on non-list key")
		}

		// 尝试对字符串进行集合操作
		_, err = cache.SMembers(ctx, "string_key")
		if err == nil {
			t.Error("Expected error when SMembers on non-set key")
		}
	})

	t.Run("Increment on Non-numeric Values", func(t *testing.T) {
		// 设置非数字值
		err := cache.Set(ctx, "non_numeric", "not_a_number", time.Hour)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// 尝试递增
		_, err = cache.Increment(ctx, "non_numeric", 1)
		if err == nil {
			t.Error("Expected error when incrementing non-numeric value")
		}
	})
}

func TestMemoryCache_HealthCheck(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestMemoryCache_Close(t *testing.T) {
	cfg := &config.CacheConfig{
		Driver:          "memory",
		MaxMemory:       100 * 1024 * 1024,
		CleanupInterval: time.Minute,
	}
	cfg.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cache, err := NewMemoryCache(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// 关闭缓存
	err = cache.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 多次关闭应该不报错
	err = cache.Close()
	if err != nil {
		t.Fatalf("Multiple close failed: %v", err)
	}
}
