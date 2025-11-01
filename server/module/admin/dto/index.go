package dto

import models "github.com/so68/core/server/database"

// LoginParams 登录参数
type LoginParams struct {
	Username string `json:"username" form:"username" validate:"required"` // 用户名
	Password string `json:"password" form:"password" validate:"required"` // 密码
	Code     string `json:"code" form:"code"`                             // 验证码
}

// LoginResult 登录结果
type LoginResult struct {
	Info  *models.Admin `json:"info"`  // 管理员信息
	Token string        `json:"token"` // 令牌
}
