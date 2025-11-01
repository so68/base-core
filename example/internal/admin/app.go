package admin

import (
	"fmt"

	"github.com/so68/core"
	"github.com/so68/core/server/module/admin"
)

// NewAdminServer 创建一个 admin 服务器
func NewAdminServer(app *core.Application) error {
	adminApp := admin.NewAdminApp(app, "/admin")
	fmt.Println(adminApp)
	return nil
}
