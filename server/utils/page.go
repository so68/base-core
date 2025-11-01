package utils

// Page 分页请求参数
type Page struct {
	Page  int64  `form:"page" binding:"omitempty"`  // 页码，从1开始
	Size  int64  `form:"size" binding:"omitempty"`  // 每页数量
	Sort  string `form:"sort" binding:"omitempty"`  // 排序字段
	Order string `form:"order" binding:"omitempty"` // 排序方向
}

// NewDefaultPage 创建默认分页参数
func NewDefaultPage() *Page {
	return &Page{
		Page:  1,
		Size:  10,
		Sort:  "id",
		Order: "DESC",
	}
}

// 获取偏移量
func (p *Page) GetOffset() int64 {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 10
	}
	return (p.Page - 1) * p.Size
}

// 获取每页数量
func (p *Page) GetLimit() int64 {
	if p.Size <= 0 {
		return 10
	}
	return p.Size
}

// 获取排序字段
func (p *Page) GetSort() string {
	return p.Sort + " " + p.Order
}

// PageResp 分页响应结构
type PageResp struct {
	Total   int64       `json:"total"`   // 总记录数
	Pages   int64       `json:"pages"`   // 总页数
	Current int64       `json:"current"` // 当前页码
	Size    int64       `form:"size"`    // 每页数量
	Items   interface{} `json:"items"`   // 数据列表
}

// NewPageResp 创建分页响应结构
func NewPageResp(total int64, page *Page, items interface{}) *PageResp {
	if page.Page <= 0 {
		page.Page = 1
	}
	if page.Size <= 0 {
		page.Size = 10
	}
	pages := total / int64(page.Size)
	if total%int64(page.Size) != 0 {
		pages++
	}
	return &PageResp{
		Total:   total,
		Pages:   pages,
		Current: page.Page,
		Size:    page.Size,
		Items:   items,
	}
}
