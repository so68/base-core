package database

import (
	"fmt"
	"log/slog"

	"github.com/so68/core/config"
)

// Factory 数据库工厂
type Factory struct {
	logger *slog.Logger
}

// NewFactory 创建数据库工厂
func NewFactory(logger *slog.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateDatabase 根据配置创建数据库连接
func (f *Factory) CreateDatabase(cfg *config.DatabaseConfig) (Database, error) {
	switch cfg.Driver {
	case "mysql":
		return NewMySQLDatabase(cfg, f.logger)
	case "postgres":
		return NewPostgreSQLDatabase(cfg, f.logger)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// CreateDatabaseWithConfig 使用默认配置创建数据库
func (f *Factory) CreateDatabaseWithConfig(driver, host string, port int, username, password, database string) (Database, error) {
	cfg := &config.DatabaseConfig{
		Driver:   driver,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}

	// 设置默认值
	cfg.SetDefaults()

	return f.CreateDatabase(cfg)
}

// CreateMySQLDatabase 创建 MySQL 数据库连接
func (f *Factory) CreateMySQLDatabase(host string, port int, username, password, database string) (Database, error) {
	cfg := &config.DatabaseConfig{
		Driver:   "mysql",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		Charset:  "utf8mb4",
		Timezone: "Local",
	}

	// 设置默认值
	cfg.SetDefaults()

	return f.CreateDatabase(cfg)
}

// CreatePostgreSQLDatabase 创建 PostgreSQL 数据库连接
func (f *Factory) CreatePostgreSQLDatabase(host string, port int, username, password, database string) (Database, error) {
	cfg := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	// 设置默认值
	cfg.SetDefaults()

	return f.CreateDatabase(cfg)
}
