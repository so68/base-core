package config

import (
	"strconv"
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`   // 数据库驱动: mysql, postgres
	Host     string `yaml:"host"`     // 数据库主机
	Port     int    `yaml:"port"`     // 数据库端口
	Username string `yaml:"username"` // 用户名
	Password string `yaml:"password"` // 密码
	Database string `yaml:"database"` // 数据库名
	Charset  string `yaml:"charset"`  // 字符集
	Timezone string `yaml:"timezone"` // 时区
	SSLMode  string `yaml:"sslMode"`  // SSL模式 (postgres)

	// 连接池配置
	MaxOpenConns    int           `yaml:"maxOpenConns"`    // 最大打开连接数
	MaxIdleConns    int           `yaml:"maxIdleConns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"` // 连接最大生存时间
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"` // 连接最大空闲时间

	// GORM 配置
	LogLevel                                 string        `yaml:"logLevel"`                                 // 日志级别: silent, error, warn, info
	SlowThreshold                            time.Duration `yaml:"slowThreshold"`                            // 慢查询阈值
	PrepareStmt                              bool          `yaml:"prepareStmt"`                              // 是否预编译语句
	DisableForeignKeyConstraintWhenMigrating bool          `yaml:"disableForeignKeyConstraintWhenMigrating"` // 迁移时是否禁用外键约束
}

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "",
		Database: "test",
		Charset:  "utf8mb4",
		Timezone: "Local",
		SSLMode:  "disable",

		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,

		LogLevel:                                 "info",
		SlowThreshold:                            time.Second,
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "mysql":
		return c.getMySQLDSN()
	case "postgres":
		return c.getPostgresDSN()
	default:
		return ""
	}
}

// getMySQLDSN 获取 MySQL 连接字符串
func (c *DatabaseConfig) getMySQLDSN() string {
	dsn := c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" +
		strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=" + c.Charset

	if c.Timezone != "" {
		dsn += "&parseTime=True&loc=" + c.Timezone
	}

	return dsn
}

// getPostgresDSN 获取 PostgreSQL 连接字符串
func (c *DatabaseConfig) getPostgresDSN() string {
	dsn := "host=" + c.Host + " port=" + strconv.Itoa(c.Port) +
		" user=" + c.Username + " password=" + c.Password +
		" dbname=" + c.Database + " sslmode=" + c.SSLMode

	if c.Timezone != "" {
		dsn += " timezone=" + c.Timezone
	}

	return dsn
}

// SetDefaults 设置默认配置值
func (c *DatabaseConfig) SetDefaults() {
	if c.Charset == "" {
		c.Charset = "utf8mb4"
	}
	if c.Timezone == "" {
		c.Timezone = "Local"
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 100
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = time.Minute * 10
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.SlowThreshold == 0 {
		c.SlowThreshold = time.Second
	}
	// PrepareStmt 和 DisableForeignKeyConstraintWhenMigrating 使用默认值 false
}
