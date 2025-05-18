package dto

type GetChildrenByFullPathReq struct {
	User         string `json:"user" form:"user"`
	FullNamePath string `json:"full_name_path" form:"full_name_path"`
}

type GetByFullPathReq struct {
	User         string `json:"user" form:"user"`
	FullNamePath string `json:"full_name_path" form:"full_name_path"`
}
