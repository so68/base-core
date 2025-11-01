package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/cache"
	"github.com/so68/core/server/module/admin/service"
	"gorm.io/gorm"
)

// AdminHandler 管理员处理
type AdminHandler struct {
	adminService service.AdminService
}

// NewAdminHandler 创建一个管理员处理
func NewAdminHandler(logger *slog.Logger, db *gorm.DB, cache cache.Cache) *AdminHandler {
	return &AdminHandler{adminService: service.NewAdminService(db, cache, logger)}
}

// Index 管理员列表
func (h *AdminHandler) Index(c *gin.Context) {

}

// Create 创建管理员
func (h *AdminHandler) Create(c *gin.Context) {

}

// Update 更新管理员
func (h *AdminHandler) Update(c *gin.Context) {

}

// TokenUpdate 更新管理员Token
func (h *AdminHandler) TokenUpdate(c *gin.Context) {

}

// TokenPasswordUpdate 更新管理员Token密码
func (h *AdminHandler) TokenPasswordUpdate(c *gin.Context) {

}

// Delete 删除管理员
func (h *AdminHandler) Delete(c *gin.Context) {

}
