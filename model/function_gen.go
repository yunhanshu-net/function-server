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
	Enable     int    `json:"enable"`                            //是否启用，-1，1
	Status     string `json:"status"`                            //状态，生成中，待审核，已审核,失败
	Classify   string `json:"classify"`                          //分类
	Tags       string `json:"tags"`                              // 标签
	RenderType string `json:"render_type"`                       // 功能渲染类型
	Comment    string `json:"comment"`                           // 评价
	CostMill   int64  `json:"cost_mill"`                         //耗时毫秒
	FunctionID int64  `json:"function_id"`                       // 函数ID
	TreeID     int64  `json:"tree_id"`                           // 关联的树ID
	Length     int    `json:"length"`                            // 字符数，根据字符数来判断是否是复杂函数
	RunnerID   int64  `json:"runner_id"`                         //所属工作空间
}

func (f *FunctionGen) TableName() string {
	return "function_gen"
}

type FunctionGenRetry struct {
	Base

	FunctionGenID int64  `json:"function_gen_id" gorm:"column:function_gen_id;comment:关联的FunctionGen ID"`
	RetryIndex    int    `json:"retry_index" gorm:"column:retry_index;comment:重试序号"`
	OriginalCode  string `json:"original_code" gorm:"type:text;column:original_code;comment:重试前的代码"`
	ErrorMsg      string `json:"error_msg" gorm:"type:text;column:error_msg;comment:错误信息"`
	FixedCode     string `json:"fixed_code" gorm:"type:text;column:fixed_code;comment:修复后的代码"`
	Success       bool   `json:"success" gorm:"column:success;comment:是否成功"`
	FinalError    string `json:"final_error" gorm:"type:text;column:final_error;comment:最终错误（如果失败）"`
	RetryTime     int64  `json:"retry_time" gorm:"column:retry_time;comment:重试时间戳"`
	CostMill      int64  `json:"cost_mill" gorm:"column:cost_mill;comment:重试耗时毫秒"`
}

func (f *FunctionGenRetry) TableName() string {
	return "function_gen_retry"
}
