package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/so68/core/config"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 实现 gorm.Logger 接口
type GormLogger struct {
	logger               *slog.Logger
	LogLevel             gormlogger.LogLevel
	SlowThreshold        time.Duration
	IgnoreRecordNotFound bool
	EnableGormSource     bool
	Colorful             bool
	LogSQL               bool
}

// NewGormLogger 创建一个新的 GORM 日志记录器
func NewGormLogger(cfg *config.DatabaseConfig, slogLogger *slog.Logger) *GormLogger {
	var gormLogLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		gormLogLevel = gormlogger.Silent
	case "error":
		gormLogLevel = gormlogger.Error
	case "warn":
		gormLogLevel = gormlogger.Warn
	case "info":
		gormLogLevel = gormlogger.Info
	default:
		gormLogLevel = gormlogger.Info
	}

	return &GormLogger{
		logger:               slogLogger,
		LogLevel:             gormLogLevel,
		SlowThreshold:        cfg.SlowThreshold,
		IgnoreRecordNotFound: true,
		EnableGormSource:     true,
		Colorful:             true,
		LogSQL:               true,
	}
}

// LogMode 实现 gorm.Logger 接口
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 实现 gorm.Logger 接口
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 实现 gorm.Logger 接口
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 实现 gorm.Logger 接口
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 实现 gorm.Logger 接口
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建日志消息
	var msg string
	if l.EnableGormSource {
		msg = fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	} else {
		msg = fmt.Sprintf("[%.3fms] [rows:%v]", float64(elapsed.Nanoseconds())/1e6, rows)
	}

	// 根据条件记录日志
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		// 记录错误
		l.logger.Error(msg, "error", err)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		// 记录慢查询
		l.logger.Warn(msg, "slow_query", fmt.Sprintf(">%v", l.SlowThreshold))
	case l.LogSQL:
		// 记录普通 SQL，使用自定义的调用栈深度
		if pc, file, line, ok := runtime.Caller(5); ok {
			// 获取相对路径
			relFile := file
			if wd, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(wd, file); err == nil {
					relFile = rel
				}
			}
			// 获取函数名
			funcName := runtime.FuncForPC(pc).Name()
			// 添加到日志行
			msg = fmt.Sprintf("%s | %s:%d | %s", msg, relFile, line, funcName)
		}
		l.logger.Info(msg)
	}
}
