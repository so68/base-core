package service

import (
	"errors"
	"log/slog"
	"regexp"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/so68/core/cache"
	"github.com/so68/core/server/module/admin/repo"
	"github.com/so68/core/server/utils"
	"gorm.io/gorm"
)

var enforcer *casbin.Enforcer

const (
	RoleSuperAdmin = "超级管理员"
	RoleMerchant   = "商户管理员"
	RoleAgent      = "代理管理员"
)

// CasbinService 权限服务
type CasbinService interface {
	// GetContextRole 获取上下文角色
	// @param c 上下文
	// @return string 角色
	// @return error 错误
	GetContextRole(c *gin.Context) (string, error)
	// AddPolicy 添加策略
	// @param role 角色
	// @param path 路径
	// @param method 方法
	// @return error 错误
	AddPolicy(role, path, method string) error
	// AddRoleInheritance 添加角色继承关系
	// @param role 角色
	// @param inheritedRole 继承角色
	// @return error 错误
	AddRoleInheritance(role, inheritedRole string) error
	// HasRoleInheritancesEnforce 检查角色是否具有继承权限
	// @param role 角色
	// @param path 路径
	// @param method 方法
	// @return bool 是否具有继承权限
	HasRoleInheritancesEnforce(role string, path string, method string) bool
}

// CasbinServiceImpl 权限服务实现
type CasbinServiceImpl struct {
	db        *gorm.DB
	cache     cache.Cache
	logger    *slog.Logger
	adminRepo repo.AdminRepo
}

// NewCasbinService 创建一个权限服务
func NewCasbinService(db *gorm.DB, cache cache.Cache, logger *slog.Logger) CasbinService {
	if enforcer == nil {
		adapter, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			panic("创建 GORM 适配器失败: " + err.Error())
		}

		// 使用代码定义 RBAC 模型，不使用配置文件
		m := model.NewModel()
		m.AddDef("r", "r", "sub, obj, act")
		m.AddDef("p", "p", "sub, obj, act")
		m.AddDef("g", "g", "_, _")
		m.AddDef("e", "e", "some(where (p.eft == allow))")
		m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

		enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil {
			panic("创建 Casbin 失败: " + err.Error())
		}

		// 加载数据库中的策略（如果有）
		err = enforcer.LoadPolicy()
		if err != nil {
			panic("加载 Casbin 策略失败: " + err.Error())
		}

		// 检查超级管理员策略是否已存在，不存在则添加
		adminRoles := []string{RoleSuperAdmin, RoleMerchant, RoleAgent}
		for _, role := range adminRoles {
			exists, _ := enforcer.HasPolicy(role, "role", "read")
			if !exists {
				_, err = enforcer.AddPolicy(role, "role", "read")
				if err != nil {
					panic("添加超级管理员失败: " + err.Error())
				}
			}
		}
		// 显式保存到数据库（虽然 GORM 适配器会自动保存，但这里确保持久化）
		err = enforcer.SavePolicy()
		if err != nil {
			panic("保存超级管理员策略失败: " + err.Error())
		}
	}
	return &CasbinServiceImpl{db: db, cache: cache, logger: logger, adminRepo: repo.NewAdminRepo()}
}

// GetContextRole 获取上下文角色
func (s *CasbinServiceImpl) GetContextRole(c *gin.Context) (string, error) {
	adminID := utils.GetContextUserID(c)
	ctx := c.Request.Context()
	admin, err := s.adminRepo.Find(ctx, utils.NewGormBuilderFind(ctx, s.db, "id", adminID))
	if err != nil {
		return "", err
	}
	return admin.Role, nil
}

// AddPolicy 添加策略
func (s *CasbinServiceImpl) AddPolicy(role, path, method string) error {
	if path == "" || method == "" || role == "" {
		return errors.New("路径、方法和角色是必需的: " + path + ", " + method + ", " + role)
	}

	// /:xxx 转换成 /*
	re := regexp.MustCompile(`/:(\w+)`)
	path = re.ReplaceAllString(path, "/*")

	// 检查策略是否已存在
	hasPolicy, _ := enforcer.HasPolicy(role, path, method)
	if hasPolicy {
		return nil
	}

	// 添加策略
	_, err := enforcer.AddPolicy(role, path, method)
	if err != nil {
		return err
	}
	// 显式保存到数据库（虽然 GORM 适配器会自动保存，但这里确保持久化）
	return enforcer.SavePolicy()
}

// AddRoleInheritance 添加角色继承关系
func (s *CasbinServiceImpl) AddRoleInheritance(role, inheritedRole string) error {
	if role == "" || inheritedRole == "" {
		return errors.New("角色和继承角色是必需的: " + role + ", " + inheritedRole)
	}

	// 添加角色继承关系
	_, err := enforcer.AddGroupingPolicy(role, inheritedRole)
	if err != nil {
		return err
	}
	return enforcer.SavePolicy()
}

// HasRoleInheritancesEnforce 检查角色是否具有继承权限
func (s *CasbinServiceImpl) HasRoleInheritancesEnforce(role string, path string, method string) bool {
	ok, err := enforcer.Enforce(role, path, method)
	if err != nil {
		s.logger.Warn("检查角色是否具有继承权限失败", "error", err)
		return false
	}
	return ok
}
