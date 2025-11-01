package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/so68/core/cache"
	"github.com/so68/core/server/database"
	"github.com/so68/core/server/module/admin/dto"
	"github.com/so68/core/server/module/admin/repo"
	"github.com/so68/core/server/utils"
	"gorm.io/gorm"
)

// IndexService 首页服务
type IndexService interface {
	// Login 管理员登陆
	// @param ctx 上下文
	// @param loginIP 登录IP
	// @param bodyParams 登录参数
	// @return *dto.LoginResult 登录结果
	// @return error 错误
	Login(ctx context.Context, loginIP string, bodyParams *dto.LoginParams) (*dto.LoginResult, error)
}

// IndexServiceImpl 首页服务实现
type IndexServiceImpl struct {
	jwt       *utils.JWT
	db        *gorm.DB
	cache     cache.Cache
	logger    *slog.Logger
	adminRepo repo.AdminRepo
}

// NewIndexService 创建一个首页服务
func NewIndexService(logger *slog.Logger, db *gorm.DB, cache cache.Cache, jwt *utils.JWT) IndexService {
	return &IndexServiceImpl{
		jwt:       jwt,
		db:        db,
		cache:     cache,
		logger:    logger,
		adminRepo: repo.NewAdminRepo(),
	}
}

// Login 管理员登陆
func (s *IndexServiceImpl) Login(ctx context.Context, loginIP string, bodyParams *dto.LoginParams) (*dto.LoginResult, error) {
	// 查询管理员
	admin, err := s.adminRepo.Find(ctx, utils.NewGormBuilderFind(ctx, s.db, "username", bodyParams.Username))
	if err != nil {
		return nil, fmt.Errorf("查询管理员失败: %w", err)
	}

	// 检查管理员是否锁定
	if admin.IsLocked() {
		return nil, fmt.Errorf("管理员已锁定,请联系管理员解锁! 锁定截止时间: %s", admin.LockedUntil.Format(time.DateTime))
	}

	// 检查管理员密码是否正确
	if !admin.CompareHashAndPassword(bodyParams.Password) {
		updateAdmin := &database.Admin{}
		updateAdmin.FailedLoginAttempts += 1

		// 检查管理员是否多次登录失败, 锁定 5 分钟
		if admin.FailedLoginAttempts >= 5 {
			updateAdmin.Status = database.AdminStatusLocked
			updateAdmin.LockedUntil = time.Now().Add(time.Minute * 5)
		}
		s.adminRepo.Update(ctx, utils.NewGormBuilderFind(ctx, s.db, "id", admin.ID), updateAdmin)
		return nil, fmt.Errorf("账号或密码错误, 请重新输入! 剩余 %d 次机会", 5-admin.FailedLoginAttempts)
	}

	// 是否开启Google Authenticator 验证
	if admin.IsMFAEnabled && !admin.VerifyGoogleAuthCode(bodyParams.Code) {
		return nil, errors.New("-Google Authenticator 验证失败, 请重新输入")
	}

	// 更新管理员登录信息
	updateAdmin := &database.Admin{LastLoginAt: time.Now(), LastLoginIP: loginIP, FailedLoginAttempts: 0}
	s.adminRepo.Update(ctx, utils.NewGormBuilderFind(ctx, s.db, "id", admin.ID), updateAdmin)

	// 返回登陆成功数据
	return &dto.LoginResult{Info: admin, Token: s.jwt.GenerateToken(admin.ID, loginIP)}, nil
}
