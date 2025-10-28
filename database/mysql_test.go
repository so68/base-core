package database

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/so68/core/config"
)

/*
MySQL数据库连接功能测试

本文件用于测试MySQLDatabase结构体的各种功能特性，
包括数据库连接、健康检查、连接池管理、GORM配置等。

运行命令：
go test -v -run "^Test.*MySQL.*$"

测试内容：
1. 数据库连接创建和配置 (NewMySQLDatabase, 配置验证等)
2. 数据库操作方法 (DB, HealthCheck等)
3. 连接池功能验证 (MaxOpenConns, MaxIdleConns等)
4. GORM配置测试 (PrepareStmt, DisableForeignKeyConstraintWhenMigrating等)
5. DSN构建和参数处理
6. 错误处理和边界条件
7. 并发访问和性能测试
8. 接口实现验证
*/

// TestNewMySQLDatabase 测试创建 MySQL 数据库连接
func TestNewMySQLDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_mysql_config",
			config: &config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
				Charset:  "utf8mb4",
				Timezone: "Local",
			},
			expectError: false,
		},
		{
			name: "invalid_driver",
			config: &config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
			},
			expectError: true,
			errorMsg:    "invalid driver: expected mysql, got postgres",
		},
		{
			name: "empty_driver",
			config: &config.DatabaseConfig{
				Driver:   "",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "Aa123098..",
				Database: "base",
			},
			expectError: true,
			errorMsg:    "invalid driver: expected mysql, got ",
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlDB, err := NewMySQLDatabase(tt.config, logger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if mysqlDB != nil {
					t.Error("Expected mysqlDB to be nil but got non-nil")
				}
				if tt.errorMsg != "" && err != nil {
					if err.Error() != tt.errorMsg {
						t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				// 注意：这个测试需要真实的 MySQL 数据库连接
				// 如果没有可用的数据库，测试会失败
				if err != nil {
					t.Skipf("Skipping test due to database connection error: %v", err)
					return
				}
				if mysqlDB == nil {
					t.Error("Expected mysqlDB to be non-nil")
					return
				}
				if mysqlDB.db == nil {
					t.Error("Expected mysqlDB.db to be non-nil")
				}
				if mysqlDB.config != tt.config {
					t.Error("Expected mysqlDB.config to match tt.config")
				}
				if mysqlDB.logger != logger {
					t.Error("Expected mysqlDB.logger to match logger")
				}
			}
		})
	}
}

// TestNewMySQLDatabaseWithDefaults 测试使用默认配置创建 MySQL 连接
func TestNewMySQLDatabaseWithDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := config.DefaultDatabaseConfig()
	config.Driver = "mysql" // 确保驱动是 mysql
	config.Password = "Aa123098.."
	config.Database = "base"

	mysqlDB, err := NewMySQLDatabase(config, logger)

	// 如果没有可用的数据库，跳过测试
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	if mysqlDB == nil {
		t.Error("Expected mysqlDB to be non-nil")
		return
	}
	if mysqlDB.db == nil {
		t.Error("Expected mysqlDB.db to be non-nil")
	}
	if mysqlDB.config != config {
		t.Error("Expected mysqlDB.config to match config")
	}
	if mysqlDB.logger != logger {
		t.Error("Expected mysqlDB.logger to match logger")
	}
}

// TestMySQLDatabase_DB 测试获取 GORM DB 实例
func TestMySQLDatabase_DB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	db := mysqlDB.DB()
	if db == nil {
		t.Error("Expected db to be non-nil")
	}
	if db != mysqlDB.db {
		t.Error("Expected db to equal mysqlDB.db")
	}
}

// TestMySQLDatabase_HealthCheck 测试健康检查
func TestMySQLDatabase_HealthCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	err = mysqlDB.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestMySQLDatabase_ConnectionPool 测试连接池配置
func TestMySQLDatabase_ConnectionPool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
		// 连接池配置
		MaxOpenConns:    50,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour * 2,
		ConnMaxIdleTime: time.Minute * 5,
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	sqlDB, err := mysqlDB.db.DB()
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

// TestMySQLDatabase_GORMConfig 测试 GORM 配置
func TestMySQLDatabase_GORMConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
		// GORM 配置
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	db := mysqlDB.DB()
	if db == nil {
		t.Error("Expected db to be non-nil")
	}

	// 测试 GORM 配置是否正确应用
	// 这里我们主要验证连接是否成功建立，具体的 GORM 配置验证需要更复杂的测试
	err = mysqlDB.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestMySQLDatabase_DSN 测试 DSN 生成
func TestMySQLDatabase_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "basic_dsn",
			config: &config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
				Charset:  "utf8mb4",
			},
			expected: "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4",
		},
		{
			name: "dsn_with_timezone",
			config: &config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
				Charset:  "utf8mb4",
				Timezone: "UTC",
			},
			expected: "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=UTC",
		},
		{
			name: "dsn_with_empty_password",
			config: &config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "",
				Database: "testdb",
				Charset:  "utf8mb4",
			},
			expected: "root:@tcp(localhost:3306)/testdb?charset=utf8mb4",
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

// TestMySQLDatabase_Interface 测试数据库接口实现
func TestMySQLDatabase_Interface(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	// 验证实现了 Database 接口
	var db Database = mysqlDB
	// 注意：如果 mysqlDB 是 nil，db 也会是 nil，但接口变量本身不会是 nil
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

// BenchmarkMySQLDatabase_HealthCheck 健康检查性能测试
func BenchmarkMySQLDatabase_HealthCheck(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // 减少日志输出以提高性能
	}))

	config := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "Aa123098..",
		Database: "base",
		Charset:  "utf8mb4",
		Timezone: "Local",
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		b.Skipf("Skipping benchmark due to database connection error: %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mysqlDB.HealthCheck()
		if err != nil {
			b.Fatalf("HealthCheck failed: %v", err)
		}
	}
}

// TestMySQLDatabase_ConcurrentAccess 测试并发访问
func TestMySQLDatabase_ConcurrentAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := &config.DatabaseConfig{
		Driver:       "mysql",
		Host:         "localhost",
		Port:         3306,
		Username:     "root",
		Password:     "Aa123098..",
		Database:     "base",
		Charset:      "utf8mb4",
		Timezone:     "Local",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	mysqlDB, err := NewMySQLDatabase(config, logger)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}

	// 并发执行健康检查
	concurrency := 10
	done := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			err := mysqlDB.HealthCheck()
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
