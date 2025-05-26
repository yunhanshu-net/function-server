package dto

import "github.com/yunhanshu-net/pkg/query"

type GetChildrenByFullPathReq struct {
	User         string `json:"user" form:"user"`
	FullNamePath string `json:"full_name_path" form:"full_name_path"`
}

type GetByFullPathReq struct {
	User         string `json:"user" form:"user"`
	FullNamePath string `json:"full_name_path" form:"full_name_path"`
}

type Search struct {
	query.PageInfoReq
	Type string `json:"type" form:"type"`
}

type UserWorkCountList struct {
	Keyword string `json:"keyword" form:"keyword"`
}
