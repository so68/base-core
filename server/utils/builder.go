package utils

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type GormBuilderWhereOperator string

const (
	GormBuilderWhereOperatorEqual            GormBuilderWhereOperator = "="           // 等于
	GormBuilderWhereOperatorNotEqual         GormBuilderWhereOperator = "!="          // 不等于
	GormBuilderWhereOperatorGreaterThan      GormBuilderWhereOperator = ">"           // 大于
	GormBuilderWhereOperatorGreaterThanEqual GormBuilderWhereOperator = ">="          // 大于等于
	GormBuilderWhereOperatorLessThan         GormBuilderWhereOperator = "<"           // 小于
	GormBuilderWhereOperatorLessThanEqual    GormBuilderWhereOperator = "<="          // 小于等于
	GormBuilderWhereOperatorLike             GormBuilderWhereOperator = "LIKE"        // 模糊匹配
	GormBuilderWhereOperatorNotLike          GormBuilderWhereOperator = "NOT LIKE"    // 不模糊匹配
	GormBuilderWhereOperatorIn               GormBuilderWhereOperator = "IN"          // 在列表中
	GormBuilderWhereOperatorNotIn            GormBuilderWhereOperator = "NOT IN"      // 不在列表中
	GormBuilderWhereOperatorBetween          GormBuilderWhereOperator = "BETWEEN"     // 在范围内
	GormBuilderWhereOperatorIsNull           GormBuilderWhereOperator = "IS NULL"     // 为空
	GormBuilderWhereOperatorIsNotNull        GormBuilderWhereOperator = "IS NOT NULL" // 不为空
)

// GormBuilderWhere GORM 构建器条件
type GormBuilderWhere struct {
	Operator GormBuilderWhereOperator
	Field    string
	Value    interface{}
}

// GormBuilderJoin GORM 构建器连接
type GormBuilderJoin struct {
	Query string        // 查询
	Args  []interface{} // 参数
}

// GormBuilder GORM 构建器
type GormBuilder struct {
	ctx  context.Context
	db   *gorm.DB
	Page *Page

	selects  []string            // 选择字段
	preloads []string            // 预加载
	joins    []*GormBuilderJoin  // 连接
	wheres   []*GormBuilderWhere // 条件
	groups   []string            // 分组
}

// NewGormBuilder 创建 GORM 构建器
func NewGormBuilder(ctx context.Context, db *gorm.DB) *GormBuilder {
	return &GormBuilder{
		db:       db,
		Page:     &Page{},
		selects:  make([]string, 0),
		preloads: make([]string, 0),
		joins:    make([]*GormBuilderJoin, 0),
		wheres:   make([]*GormBuilderWhere, 0),
		groups:   make([]string, 0),
	}
}

// NewGormBuilderFind 创建查询构建器
func NewGormBuilderFind(ctx context.Context, db *gorm.DB, field string, value interface{}) *GormBuilder {
	return NewGormBuilder(ctx, db).WhereEqual(field, value)
}

// NewGormBuilderWithPage 创建分页查询构建器
func NewGormBuilderWithPage(ctx context.Context, db *gorm.DB, page *Page) *GormBuilder {
	return NewGormBuilder(ctx, db).WithPage(page)
}

// TotalCount 获取总记录数
func (b *GormBuilder) TotalCount(n int64) *GormBuilder {
	b.build().Count(&n)
	return b
}

// Find 查询数据
func (b *GormBuilder) Find(data interface{}) error {
	// 分页
	if b.Page.Size > 0 {
		b.db.Offset(int(b.Page.GetOffset())).Limit(int(b.Page.GetLimit()))
	}
	// 字段排序
	if b.Page.Sort != "" {
		b.db.Order(b.Page.GetSort())
	}
	if err := b.build().Find(data).Error; err != nil {
		return err
	}
	return nil
}

// First 查询单条数据
func (b *GormBuilder) First(data interface{}) error {
	if err := b.build().First(data).Error; err != nil {
		return err
	}
	return nil
}

// Create 创建数据
func (b *GormBuilder) Create(data interface{}) error {
	return b.db.WithContext(b.ctx).Create(data).Error
}

// Update 更新数据
func (b *GormBuilder) Update(data interface{}) error {
	return b.build().WithContext(b.ctx).Save(data).Error
}

// Delete 删除数据
func (b *GormBuilder) Delete(isScoped bool, model interface{}) error {
	if isScoped {
		return b.build().WithContext(b.ctx).Delete(model).Error
	}
	return b.build().WithContext(b.ctx).Unscoped().Delete(model).Error
}

// WithPage 设置分页参数
func (b *GormBuilder) WithPage(page *Page) *GormBuilder {
	b.Page = page
	return b
}

// Select 添加选择字段
func (b *GormBuilder) Select(fields ...string) *GormBuilder {
	b.selects = append(b.selects, fields...)
	return b
}

// Preload 添加预加载
func (b *GormBuilder) Preload(relation string) *GormBuilder {
	b.preloads = append(b.preloads, relation)
	return b
}

// Join 添加连接
func (b *GormBuilder) Join(query string, args ...interface{}) *GormBuilder {
	b.joins = append(b.joins, &GormBuilderJoin{
		Query: query,
		Args:  args,
	})
	return b
}

// Group 添加分组
func (b *GormBuilder) Group(fields ...string) *GormBuilder {
	b.groups = append(b.groups, fields...)
	return b
}

// Where 添加条件
func (b *GormBuilder) Where(operator GormBuilderWhereOperator, field string, value interface{}) *GormBuilder {
	b.wheres = append(b.wheres, &GormBuilderWhere{
		Operator: operator,
		Field:    field,
		Value:    value,
	})
	return b
}

// WhereEqual 等于
func (b *GormBuilder) WhereEqual(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorEqual, field, value)
}

// WhereNotEqual 不等于
func (b *GormBuilder) WhereNotEqual(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorNotEqual, field, value)
}

// WhereGreaterThan 大于
func (b *GormBuilder) WhereGreaterThan(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorGreaterThan, field, value)
}

// WhereGreaterThanEqual 大于等于
func (b *GormBuilder) WhereGreaterThanEqual(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorGreaterThanEqual, field, value)
}

// WhereLessThan 小于
func (b *GormBuilder) WhereLessThan(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorLessThan, field, value)
}

// WhereLessThanEqual 小于等于
func (b *GormBuilder) WhereLessThanEqual(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorLessThanEqual, field, value)
}

// WhereLike 模糊匹配
func (b *GormBuilder) WhereLike(field string, value interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorLike, field, value)
}

// WhereIn 在列表中
func (b *GormBuilder) WhereIn(field string, value []interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorIn, field, value)
}

// WhereNotIn 不在列表中
func (b *GormBuilder) WhereNotIn(field string, value []interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorNotIn, field, value)
}

// WhereBetween 在范围内
func (b *GormBuilder) WhereBetween(field string, start, end interface{}) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorBetween, field, []interface{}{start, end})
}

// WhereIsNull 为空
func (b *GormBuilder) WhereIsNull(field string) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorIsNull, field, nil)
}

// WhereIsNotNull 不为空
func (b *GormBuilder) WhereIsNotNull(field string) *GormBuilder {
	return b.Where(GormBuilderWhereOperatorIsNotNull, field, nil)
}

// Build 构建 GORM 查询
func (b *GormBuilder) build() *gorm.DB {
	// 构建选择字段
	if len(b.selects) > 0 {
		b.db = b.db.Select(strings.Join(b.selects, ","))
	}

	// 构建预加载
	if len(b.preloads) > 0 {
		for _, preload := range b.preloads {
			b.db = b.db.Preload(preload)
		}
	}

	// 构建连接
	if len(b.joins) > 0 {
		for _, join := range b.joins {
			b.db = b.db.Joins(join.Query, join.Args...)
		}
	}

	// 构建条件
	for _, where := range b.wheres {
		switch where.Operator {
		case GormBuilderWhereOperatorBetween:
			// 范围类型需要两个值 安全的类型断言
			if values, ok := where.Value.([]interface{}); ok && len(values) == 2 {
				b.db = b.db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", where.Field), values[0], values[1])
			}
		case GormBuilderWhereOperatorIsNull, GormBuilderWhereOperatorIsNotNull:
			// IS NULL 和 IS NOT NULL 不需要值
			b.db = b.db.Where(fmt.Sprintf("%s %s", where.Field, where.Operator))
		default:
			// 其他操作符使用单个值
			b.db = b.db.Where(fmt.Sprintf("%s %s ?", where.Field, where.Operator), where.Value)
		}
	}

	// 构建分组
	if len(b.groups) > 0 {
		b.db = b.db.Group(strings.Join(b.groups, ","))
	}

	return b.db
}
