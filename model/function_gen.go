package model

type FunctionGen struct {
	Base

	Message    string `json:"message" gorm:"type:varchar(5000)"` //用户需求
	Code       string `json:"code" gorm:"type:text"`             //源代码
	UpdateCode string `json:"update_code" gorm:"type:text"`      //更新后的代码
	Thinking   string `json:"thinking" gorm:"type:text"`         //思考过程
	Score      int64  `json:"score"`                             //得分
	Level      int64  `json:"level"`                             //函数复杂度：1-100
	Quality    string `json:"quality"`                           //质量，优，良，中，差
	Enable     int    `json:"enable"`                            //是否启用
	Status     string `json:"status"`                            //状态，未审核，已审核，
	Classify   string `json:"classify"`                          //分类
	Tags       string `json:"tags"`                              // 标签
	RenderType string `json:"render_type"`                       // 功能渲染类型
	Comment    string `json:"comment"`                           // 评价
	CostMill   int64  `json:"cost_mill"`                         //耗时毫秒
	FunctionID int64  `json:"function_id"`                       // 函数ID
	TreeID     int64  `json:"tree_id"`                           // 关联的树ID
	Length     int    `json:"length"`                            // 字符数，根据字符数来判断是否是复杂函数
}

func (f *FunctionGen) TableName() string {
	return "function_gen"
}
