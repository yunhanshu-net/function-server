package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/response"
)

type UserWorkCountList struct {
	User  string `json:"user"`
	Count int    `json:"count"`
}

func (api *ServiceTreeAPI) UserWorkCountList(c *gin.Context) {
	var res []UserWorkCountList
	var q = c.Query("fuzzy")
	err := db.GetDB().Model(&model.ServiceTree{}).
		Select("user, count(*) as count").
		Where("user like ? ", "%"+q+"%").Limit(20).Order("count desc").Group("user").Scan(&res).Error
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, res)
}
