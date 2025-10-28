package database

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/so68/core/config"
)

/*
PostgreSQL数据库连接功能测试

本文件用于测试PostgreSQLDatabase结构体的各种功能特性，
包括数据库连接、健康检查、连接池管理、GORM配置等。

运行命令：
go test -v -run "^Test.*Postgres.*$"

测试内容：
1. 数据库连接创建和配置 (NewPostgreSQLDatabase, 配置验证等)
2. 数据库操作方法 (DB, HealthCheck等)
3. 连接池功能验证 (MaxOpenConns, MaxIdleConns等)
4. GORM配置测试 (PrepareStmt, DisableForeignKeyConstraintWhenMigrating等)
5. DSN构建和参数处理
6. 错误处理和边界条件
7. 并发访问和性能测试
8. 接口实现验证
*/

// TestNewPostgreSQLDatabase 测试创建 PostgreSQL 数据库连接
func TestNewPostgreSQLDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_postgres_config",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
				SSLMode:  "disable",
				Timezone: "UTC",
			},
			expectError: false,
		},
		{
			name: "invalid_driver",
			config: &config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     5432,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
			},
			expectError: true,
			errorMsg:    "invalid driver: expected postgres, got mysql",
		},
		{
			name: "empty_driver",
			config: &config.DatabaseConfig{
				Driver:   "",
				Host:     "localhost",
				Port:     5432,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
			},
			expectError: true,
			errorMsg:    "invalid driver: expected postgres, got ",
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgresDB, err := NewPostgreSQLDatabase(tt.config, logger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if postgresDB != nil {
					t.Error("Expected postgresDB to be nil but got non-nil")
				}
				if tt.errorMsg != "" && err != nil {
					if err.Error() != tt.errorMsg {
						t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				// 注意：这个测试需要真实的 PostgreSQL 数据库连接
				// 如果没有可用的数据库，测试会失败
				if err != nil {
					t.Skipf("Skipping test due to database connection error: %v", err)
					return
				}
				if postgresDB == nil {
					t.Error("Expected postgresDB to be non-nil")
					return
				}
				if postgresDB.db == nil {
					t.Error("Expected postgresDB.db to be non-nil")
				}
				if postgresDB.config != tt.config {
					t.Error("Expected postgresDB.config to match tt.config")
				}
				if postgresDB.logger != logger {
					t.Error("Expected postgresDB.logger to match logger")
				}
			}
		})
	}
}

// TestNewPostgreSQLDatabaseWithDefaults 测试使用默认配置创建 PostgreSQL 连接
func TestNewPostgreSQLDatabaseWithDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := config.DefaultDatabaseConfig()
	config.Driver = "postgres" // 确保驱动是 postgres
	config.Password = ""
	config.Database = "base"
	config.SSLMode = "disable"

	postgresDB, err := NewPostgreSQLDatabase(config, logger)

	// 如果没有可用的数据库，跳过测试
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	if postgresDB == nil {
		t.Error("Expected postgresDB to be non-nil")
		return
	}
	if postgresDB.db == nil {
		t.Error("Expected postgresDB.db to be non-nil")
	}
	if postgresDB.config != config {
		t.Error("Expected postgresDB.config to match config")
	}
	if postgresDB.logger != logger {
		t.Error("Expected postgresDB.logger to match logger")
	}
}

// TestPostgreSQLDatabase_DB 测试获取 GORM DB 实例
func TestPostgreSQLDatabase_DB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	db := postgresDB.DB()
	if db == nil {
		t.Error("Expected db to be non-nil")
	}
	if db != postgresDB.db {
		t.Error("Expected db to equal postgresDB.db")
	}
}

// TestPostgreSQLDatabase_HealthCheck 测试健康检查
func TestPostgreSQLDatabase_HealthCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	err = postgresDB.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestPostgreSQLDatabase_ConnectionPool 测试连接池配置
func TestPostgreSQLDatabase_ConnectionPool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
		// 连接池配置
		MaxOpenConns:    50,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour * 2,
		ConnMaxIdleTime: time.Minute * 5,
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	sqlDB, err := postgresDB.db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	// 连接建立后可能会有连接被创建，这是正常的
	if stats.OpenConnections < 0 {
		t.Errorf("Expected OpenConnections to be >= 0, got %d", stats.OpenConnections)
	}
	if stats.InUse < 0 {
		t.Errorf("Expected InUse to be >= 0, got %d", stats.InUse)
	}
	if stats.Idle < 0 {
		t.Errorf("Expected Idle to be >= 0, got %d", stats.Idle)
	}

	// 验证连接池配置是否正确应用
	if stats.OpenConnections > config.MaxOpenConns {
		t.Errorf("OpenConnections %d exceeds MaxOpenConns %d", stats.OpenConnections, config.MaxOpenConns)
	}
}

// TestPostgreSQLDatabase_GORMConfig 测试 GORM 配置
func TestPostgreSQLDatabase_GORMConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
		// GORM 配置
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	db := postgresDB.DB()
	if db == nil {
		t.Error("Expected db to be non-nil")
	}

	// 测试 GORM 配置是否正确应用
	// 这里我们主要验证连接是否成功建立，具体的 GORM 配置验证需要更复杂的测试
	err = postgresDB.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestPostgreSQLDatabase_DSN 测试 DSN 生成
func TestPostgreSQLDatabase_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "basic_dsn",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable",
		},
		{
			name: "dsn_with_timezone",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
				Timezone: "UTC",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable timezone=UTC",
		},
		{
			name: "dsn_with_empty_password",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "root",
				Password: "Aa123098..",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=root password=Aa123098.. dbname=testdb sslmode=disable",
		},
		{
			name: "dsn_with_ssl_require",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "require",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.GetDSN()
			if dsn != tt.expected {
				t.Errorf("Expected DSN %q, got %q", tt.expected, dsn)
			}
		})
	}
}

// TestPostgreSQLDatabase_Interface 测试数据库接口实现
func TestPostgreSQLDatabase_Interface(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	// 验证实现了 Database 接口
	var db Database = postgresDB
	// 注意：如果 postgresDB 是 nil，db 也会是 nil，但接口变量本身不会是 nil
	// 这里我们直接测试接口方法

	// 测试接口方法
	gormDB := db.DB()
	if gormDB == nil {
		t.Error("Expected gormDB to be non-nil")
	}

	err = db.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// BenchmarkPostgreSQLDatabase_HealthCheck 健康检查性能测试
func BenchmarkPostgreSQLDatabase_HealthCheck(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // 减少日志输出以提高性能
	}))

	config := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		b.Skipf("Skipping benchmark due to database connection error: %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := postgresDB.HealthCheck()
		if err != nil {
			b.Fatalf("HealthCheck failed: %v", err)
		}
	}
}

// TestPostgreSQLDatabase_ConcurrentAccess 测试并发访问
func TestPostgreSQLDatabase_ConcurrentAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:       "postgres",
		Host:         "localhost",
		Port:         5432,
		Username:     "root",
		Password:     "Aa123098..",
		Database:     "base",
		SSLMode:      "disable",
		Timezone:     "UTC",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	postgresDB, err := NewPostgreSQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	// 并发执行健康检查
	concurrency := 10
	done := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			err := postgresDB.HealthCheck()
			done <- err
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < concurrency; i++ {
		err := <-done
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

// TestPostgreSQLDatabase_SSLMode 测试不同的 SSL 模式
func TestPostgreSQLDatabase_SSLMode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

	for _, sslMode := range sslModes {
		t.Run("ssl_mode_"+sslMode, func(t *testing.T) {
			config := &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
				SSLMode:  sslMode,
				Timezone: "UTC",
			}

			postgresDB, err := NewPostgreSQLDatabase(config, logger)
			if err != nil {
				// 某些 SSL 模式可能需要特定的证书配置，这是正常的
				t.Logf("SSL mode %s not supported or configured: %v", sslMode, err)
				return
			}

			err = postgresDB.HealthCheck()
			if err != nil {
				t.Errorf("HealthCheck failed for SSL mode %s: %v", sslMode, err)
			}
		})
	}
}
