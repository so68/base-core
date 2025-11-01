package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Resp 统一响应结构
type Resp struct {
	Code    int         `json:"code"`           // 业务状态码
	Message string      `json:"message"`        // 提示信息
	Data    interface{} `json:"data,omitempty"` // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Resp{
		Code:    -1,
		Message: message,
	})
}
