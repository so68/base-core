package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Database 数据库接口
type Database interface {
	// 基础操作
	DB() *gorm.DB

	// 健康检查
	HealthCheck() error

	// 连接管理
	Close(ctx context.Context) error
}

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
