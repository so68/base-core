package handler

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/cache"
	"github.com/so68/core/server/module/admin/dto"
	"github.com/so68/core/server/module/admin/service"
	"github.com/so68/core/server/utils"
	"gorm.io/gorm"
)

// IndexHandler 首页处理
type IndexHandler struct {
	maxHeaderSize int64                // 最大请求头大小
	staticPath    string               // 静态文件路径
	indexService  service.IndexService // 首页服务
}

// NewIndexHandler 创建一个首页处理
func NewIndexHandler(logger *slog.Logger, db *gorm.DB, cache cache.Cache, jwt *utils.JWT, staticPath string, maxHeaderSize int64) *IndexHandler {
	return &IndexHandler{
		maxHeaderSize: maxHeaderSize,
		staticPath:    staticPath,
		indexService:  service.NewIndexService(logger, db, cache, jwt),
	}
}

// Login 管理员登陆
func (h *IndexHandler) Login(c *gin.Context) {
	bodyParams := &dto.LoginParams{}
	if err := c.ShouldBindJSON(bodyParams); err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 管理员登陆业务处理
	result, err := h.indexService.Login(c.Request.Context(), c.ClientIP(), bodyParams)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 返回登录成功响应
	utils.Success(c, result)
}

// Upload 上传文件
func (h *IndexHandler) Upload(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 验证文件大小 (限制为500MB)
	if file.Size > h.maxHeaderSize {
		utils.Error(c, fmt.Sprintf("文件大小超过限制: %d > %d", file.Size, h.maxHeaderSize))
		return
	}

	// 生成安全的文件名
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))

	// 确保上传目录存在
	uploadDir := filepath.Join(h.staticPath, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 保存文件
	filepath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 返回文件URL
	utils.Success(c, fmt.Sprintf("/%s", filepath))
}
