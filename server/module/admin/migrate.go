package admin

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/so68/core/server/database"
	"github.com/so68/core/server/module/admin/service"
	"gorm.io/gorm"
)

// InitMigrate 初始化迁移
func InitMigrate(db *gorm.DB, logger *slog.Logger) error {
	var nums int64
	// 迁移管理员表
	if err := db.AutoMigrate(&database.Admin{}); err != nil {
		return fmt.Errorf("迁移管理员表失败: %w", err)
	}
	// 载入管理员数据
	if err := db.Model(&database.Admin{}).Count(&nums).Error; err == nil && nums == 0 {
		// 载入管理员数据
		admin := &database.Admin{
			Username:     "superadmin",
			Nickname:     "超级管理员",
			Email:        "superadmin@example.com",
			Telephone:    "12345678901",
			PasswordHash: "Aa123098.78",
			Type:         database.AdminTypeSuper,
			Role:         service.RoleSuperAdmin,
			LastLoginAt:  time.Now(),
			LastLoginIP:  "127.0.0.1",
			Data: database.AdminData{
				WhiteList: "127.0.0.1",
			},
		}
		if err := db.Create(admin).Error; err != nil {
			return fmt.Errorf("载入管理员数据失败: %w", err)
		}
	}
	return nil
}
