package database

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/so68/core/config"
)

// MySQLDatabase MySQL 数据库实现
type MySQLDatabase struct {
	db     *gorm.DB
	config *config.DatabaseConfig
	logger *slog.Logger
}

// NewMySQLDatabase 创建 MySQL 数据库连接
func NewMySQLDatabase(cfg *config.DatabaseConfig, slogLogger *slog.Logger) (*MySQLDatabase, error) {
	if cfg.Driver != "mysql" {
		return nil, fmt.Errorf("invalid driver: expected mysql, got %s", cfg.Driver)
	}

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger:                                   NewGormLogger(cfg, slogLogger),
		PrepareStmt:                              cfg.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: cfg.DisableForeignKeyConstraintWhenMigrating,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 获取底层 sql.DB 进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	mysqlDB := &MySQLDatabase{
		db:     db,
		config: cfg,
		logger: slogLogger,
	}

	mysqlDB.logger.Info("MySQL database connected successfully",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
		slog.Int("max_open_conns", cfg.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	return mysqlDB, nil
}

// DB 获取 GORM DB 实例
func (m *MySQLDatabase) DB() *gorm.DB {
	return m.db
}

// HealthCheck 健康检查
func (m *MySQLDatabase) HealthCheck() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}

	// 检查连接池状态
	stats := sqlDB.Stats()
	m.logger.Info("MySQL connection pool stats",
		slog.Int("open_connections", stats.OpenConnections),
		slog.Int("in_use", stats.InUse),
		slog.Int("idle", stats.Idle),
		slog.Int64("wait_count", stats.WaitCount),
		slog.Duration("wait_duration", stats.WaitDuration),
	)

	return sqlDB.Ping()
}

// Close 关闭数据库连接
func (m *MySQLDatabase) Close(ctx context.Context) error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	// ctx 暂不用于 gorm 原生 close，可用于未来扩展
	return sqlDB.Close()
}
