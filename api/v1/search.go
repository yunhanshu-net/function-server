package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/pkg/query"
)

func (api *ServiceTreeAPI) Search(c *gin.Context) {
	r := dto.Search{}
	err := c.ShouldBindQuery(&r)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	var list []*model.ServiceTree

	// 创建基础查询对象
	baseDB := db.GetDB().Model(&model.ServiceTree{})

	// 添加自定义查询条件（keyword和type）
	if r.Keyword != "" && r.Type != "user" {
		r.Keyword = strings.ReplaceAll(r.Keyword, ".", "/")
		keyword := "%" + r.Keyword + "%"
		baseDB = baseDB.Where(
			baseDB.Where("name LIKE ?", keyword).
				Or("full_name_path LIKE ?", keyword).
				Or("title LIKE ?", keyword).
				Or("description LIKE ?", keyword),
		)
	}

	if r.Type != "" {
		if r.Type == "workspace" {
			baseDB = baseDB.Where("parent_id = ? AND type = ?", 0, "package")
		} else {
			if r.Type == "user" {
				baseDB = baseDB.Where("user = ?", r.Keyword)
			} else {
				baseDB = baseDB.Where("type = ?", r.Type)
			}
		}
	}

	// 使用AutoPaginateTable处理分页和通用查询条件（eq, like, in等）
	res, err := query.AutoPaginateTable(c, baseDB, &model.ServiceTree{}, &list, &r.PageInfoReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, res)
}
