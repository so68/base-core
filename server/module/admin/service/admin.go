package service

import (
	"log/slog"

	"github.com/so68/core/cache"
	"github.com/so68/core/server/module/admin/repo"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService interface {
}

// AdminServiceImpl 管理员服务实现
type AdminServiceImpl struct {
	db        *gorm.DB
	cache     cache.Cache
	logger    *slog.Logger
	adminRepo repo.AdminRepo
}

// NewAdminService 创建一个管理员服务
func NewAdminService(db *gorm.DB, cache cache.Cache, logger *slog.Logger) AdminService {
	return &AdminServiceImpl{
		db:        db,
		cache:     cache,
		logger:    logger,
		adminRepo: repo.NewAdminRepo(),
	}
}
