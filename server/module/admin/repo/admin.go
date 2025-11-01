package repo

import (
	"context"

	models "github.com/so68/core/server/database"
	"github.com/so68/core/server/utils"
)

// AdminRepo 管理员数据操作
type AdminRepo interface {
	// Find 构建查询
	Find(ctx context.Context, builder *utils.GormBuilder) (*models.Admin, error)
	// FindList 构建查询列表
	FindList(ctx context.Context, builder *utils.GormBuilder) ([]*models.Admin, error)
	// FindListWithPage 查询分页列表
	FindListWithPage(ctx context.Context, builder *utils.GormBuilder) (*utils.PageResp, error)
	// Create 创建管理员
	Create(ctx context.Context, builder *utils.GormBuilder, admin *models.Admin) error
	// Update 更新管理员
	Update(ctx context.Context, builder *utils.GormBuilder, admin *models.Admin) error
	// Delete 删除管理员
	Delete(ctx context.Context, builder *utils.GormBuilder, isScoped bool, id uint) error
}

// AdminRepoImpl 管理员数据操作实现
type AdminRepoImpl struct {
}

// NewAdminRepo 创建一个管理员数据操作
func NewAdminRepo() AdminRepo {
	return &AdminRepoImpl{}
}

// Find 查询单条数据
func (r *AdminRepoImpl) Find(ctx context.Context, builder *utils.GormBuilder) (*models.Admin, error) {
	var admin models.Admin
	if err := builder.First(&admin); err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindList 构建查询列表
func (r *AdminRepoImpl) FindList(ctx context.Context, builder *utils.GormBuilder) ([]*models.Admin, error) {
	var admins []*models.Admin
	if err := builder.Find(&admins); err != nil {
		return nil, err
	}
	return admins, nil
}

// FindListWithPage 构建查询分页
func (r *AdminRepoImpl) FindListWithPage(ctx context.Context, builder *utils.GormBuilder) (*utils.PageResp, error) {
	var admins []*models.Admin
	var total int64
	if err := builder.TotalCount(total).Find(&admins); err != nil {
		return nil, err
	}
	return utils.NewPageResp(total, builder.Page, admins), nil
}

// Create 创建管理员
func (r *AdminRepoImpl) Create(ctx context.Context, builder *utils.GormBuilder, admin *models.Admin) error {
	return builder.Create(admin)
}

// Update 更新管理员
func (r *AdminRepoImpl) Update(ctx context.Context, builder *utils.GormBuilder, admin *models.Admin) error {
	return builder.Update(admin)
}

// Delete 删除管理员
func (r *AdminRepoImpl) Delete(ctx context.Context, builder *utils.GormBuilder, isScoped bool, id uint) error {
	return builder.WhereEqual("id", id).Delete(isScoped, &models.Admin{})
}
