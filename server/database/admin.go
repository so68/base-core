package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/so68/core/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	AdminStatusEnabled  int8 = 1 // 启用
	AdminStatusDisabled int8 = 2 // 禁用
	AdminStatusLocked   int8 = 3 // 锁定

	// 管理员层级
	AdminTypeSuper    int8 = 1 // 超级管理员
	AdminTypeMerchant int8 = 2 // 商户管理员
	AdminTypeAgent    int8 = 3 // 代理管理员
)

// Admin 管理员/代理
type Admin struct {
	database.BaseModel

	// 用户名，唯一标识
	Username string `gorm:"type:varchar(255);uniqueIndex;not null;comment:'用户名'" json:"username"`
	// 邮箱地址，可选
	Email string `gorm:"type:varchar(255);default:null;uniqueIndex;comment:'邮箱'" json:"email"`
	// 手机号码，可选
	Telephone string `gorm:"type:varchar(255);default:null;uniqueIndex;comment:'手机号'" json:"telephone"`
	// 密码哈希值，不返回给前端
	PasswordHash string `gorm:"type:varchar(255);comment:'密码哈希'" json:"-"`
	// 安全码
	SecurityKey string `gorm:"type:varchar(255);uniqueIndex;comment:'安全码'" json:"security_key"`
	// 昵称，可选
	Nickname string `gorm:"type:varchar(100);not null;comment:'昵称'" json:"nickname"`
	// 头像URL地址，可选
	Avatar string `gorm:"type:varchar(255);comment:'头像URL'" json:"avatar"`
	// 账户状态：1-启用，2-禁用，3-锁定
	Status int8 `gorm:"default:1;not null;comment:'状态(1:启用,2:禁用,3:锁定)" json:"status"`
	// 账户锁定截止时间
	LockedUntil time.Time `gorm:"comment:'锁定截止时间'" json:"locked_until"`
	// 登录失败次数
	FailedLoginAttempts int8 `gorm:"default:0;comment:'登录失败次数'" json:"-"`
	// 最后登录时间
	LastLoginAt time.Time `gorm:"comment:'最后登录时间'" json:"last_login_at"`
	// 最后登录IP地址
	LastLoginIP string `gorm:"type:varchar(255);comment:'最后登录IP'" json:"last_login_ip"`
	// 最后密码修改时间
	PasswordChangedAt time.Time `gorm:"comment:'最后密码修改时间'" json:"password_changed_at"`
	// 管理员层级：1-超级管理员，2-商户管理员，3-代理管理员
	Type int8 `gorm:"default:1;not null;comment:'管理员层级(1超管)" json:"type"`
	// 角色
	Role string `gorm:"type:varchar(50);comment:'角色'" json:"role"`
	// 上级管理员ID
	ParentID uint `gorm:"index;default:null;comment:'上级管理员ID'" json:"parent_id"`
	// 管理金额/业绩/余额
	Amount float64 `gorm:"type:decimal(18,2);comment:'管理金额/业绩/余额'" json:"amount"`
	// 是否启用MFA双因素认证
	IsMFAEnabled bool `gorm:"not null;default:false;comment:'是否启用MFA'" json:"is_mfa_enabled"`
	// MFA密钥哈希或加密值
	MFASecret string `gorm:"type:varchar(255);comment:'MFA密钥哈希或加密值'" json:"mfa_secret"`
	// 客服链接
	ChatURL string `gorm:"type:varchar(255);comment:'客服链接'" json:"chat_url"`
	// 数据
	Data AdminData `gorm:"type:json;comment:'数据'" json:"data"`

	// 上级管理员关联
	Parent *Admin `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
	// 下级管理员/代理列表
	Subordinates []*Admin `gorm:"foreignKey:ParentID;references:ID" json:"subordinates"`
}

// AdminData 管理员数据
type AdminData struct {
	// 白名单 - 用 逗号 分隔
	WhiteList string `json:"white_list" form:"white_list" views:"label:白名单;type:textarea;"`
}

// Value 实现 driver.Valuer 接口
func (d AdminData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan 实现 sql.Scanner 接口
func (d *AdminData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("AdminData: Scan source is not []byte")
	}
	return json.Unmarshal(bytes, d)
}

// BeforeCreate GORM 钩子：创建前验证
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	// 哈希初始化密码
	if err := a.GeneratePasswordHash(a.PasswordHash); err != nil {
		return err
	}

	// 生成google 验证器密钥
	a.GenerateGoogleAuthSecret()
	return nil
}

// CompareHashAndPassword 比较密码哈希值
func (a *Admin) CompareHashAndPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) == nil
}

// GeneratePasswordHash 生成密码哈希值
func (a *Admin) GeneratePasswordHash(password string) error {
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return fmt.Errorf("管理员密码哈希失败: %v", hashErr)
	}
	a.PasswordHash = string(hashedPassword)
	return nil
}

// VerifyGoogleAuthCode 验证 Google Authenticator 验证码
func (a *Admin) VerifyGoogleAuthCode(code string) bool {
	// 使用 Google Authenticator 库验证验证码
	valid, err := totp.ValidateCustom(code, a.MFASecret, time.Now(), totp.ValidateOpts{
		Period: 30, // 30秒有效期
		Skew:   1,  // 允许前后1个时间窗口
		Digits: 6,  // 6位数字
	})

	if err != nil {
		return false
	}

	return valid
}

// GenerateGoogleAuthSecret 生成 Google Authenticator 密钥
func (a *Admin) GenerateGoogleAuthSecret() error {
	// 生成密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Superadmin", // 发行者名称
		AccountName: a.Username,   // 账户名称
		SecretSize:  20,           // 密钥长度
	})
	if err != nil {
		return fmt.Errorf("生成Google Authenticator密钥失败: %w", err)
	}

	// 保存密钥
	a.MFASecret = key.Secret()
	return nil
}

// GetGoogleAuthQRCode 获取 Google Authenticator 二维码内容
func (a *Admin) GetGoogleAuthQRCode() (string, error) {
	if a.MFASecret == "" {
		return "", errors.New("MFA密钥未设置")
	}

	// 生成二维码内容
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Admin System",
		AccountName: a.Username,
		Secret:      []byte(a.MFASecret),
	})
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %v", err)
	}

	return key.Secret(), nil
}

// IsLocked 检查管理员账户是否被锁定
func (a *Admin) IsLocked() bool {
	if a.Status != AdminStatusLocked {
		return false
	}

	// 如果锁定时间已过，则视为未锁定
	if a.LockedUntil.Before(time.Now()) {
		return false
	}

	return true
}
