package base

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

// Paginated 分页结果结构体
type Paginated[T any] struct {
	Items       T     `json:"items"`        // 分页数据
	CurrentPage int   `json:"current_page"` // 当前页码
	TotalCount  int64 `json:"total_count"`  // 总数据量
	TotalPages  int   `json:"total_pages"`  // 总页数
	PageSize    int   `json:"page_size"`    // 每页数量
}

// PageInfoReq 分页参数结构体
type PageInfoReq struct {
	Page     int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"omitempty,min=1"`
	Sorts    string `json:"sorts" form:"sorts"`

	// 查询条件
	Eq   string `form:"eq"`   // 格式：field1,value1,field2,value2
	Like string `form:"like"` // 格式：field1,value1,field2,value2
	In   string `form:"in"`   // 格式：field1,value1,value2,field2,value3,value4
	Gt   string `form:"gt"`   // 格式：field1,value1,field2,value2
	Gte  string `form:"gte"`  // 格式：field1,value1,field2,value2
	Lt   string `form:"lt"`   // 格式：field1,value1,field2,value2
	Lte  string `form:"lte"`  // 格式：field1,value1,field2,value2
}

// QueryConfig 查询配置
type QueryConfig struct {
	Fields    map[string][]string // 字段名 -> 允许的操作符列表（白名单）
	Blacklist map[string]struct{} // 不允许查询的字段（黑名单）
}

// NewQueryConfig 创建查询配置
func NewQueryConfig() *QueryConfig {
	return &QueryConfig{
		Fields:    make(map[string][]string),
		Blacklist: make(map[string]struct{}),
	}
}

// AllowField 允许字段查询
func (c *QueryConfig) AllowField(field string, operators ...string) {
	c.Fields[field] = operators
}

// DenyField 禁止字段查询
func (c *QueryConfig) DenyField(field string) {
	c.Blacklist[field] = struct{}{}
}

// GetLimit 获取分页大小，支持默认值
func (i *PageInfoReq) GetLimit(defaultSize ...int) int {
	if i.PageSize <= 0 {
		if len(defaultSize) > 0 {
			return defaultSize[0]
		}
		return 20
	}
	return i.PageSize
}

// GetOffset 获取分页偏移量
func (i *PageInfoReq) GetOffset() int {
	if i.Page < 1 {
		i.Page = 1
	}
	return (i.Page - 1) * i.GetLimit()
}

// SafeColumn 检查列名是否安全（防SQL注入）
func SafeColumn(column string) bool {
	for _, c := range column {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// ParseSortFields 解析排序字段字符串
func ParseSortFields(sortStr string) ([]string, error) {
	if sortStr == "" {
		return nil, nil
	}

	parts := strings.Split(sortStr, ",")
	if len(parts)%2 != 0 {
		return nil, fmt.Errorf("排序字段格式错误：字段名和排序方向必须成对出现")
	}

	var sortFields []string
	for i := 0; i < len(parts); i += 2 {
		field := strings.TrimSpace(parts[i])
		order := strings.TrimSpace(parts[i+1])

		if !SafeColumn(field) {
			return nil, fmt.Errorf("无效的排序字段名：%s", field)
		}

		order = strings.ToUpper(order)
		if order != "ASC" && order != "DESC" {
			return nil, fmt.Errorf("无效的排序方向：%s", order)
		}

		sortFields = append(sortFields, fmt.Sprintf("%s %s", field, order))
	}

	return sortFields, nil
}

// GetSorts 获取排序SQL
func (i *PageInfoReq) GetSorts() string {
	sortFields, err := ParseSortFields(i.Sorts)
	if err != nil || len(sortFields) == 0 {
		return ""
	}
	return strings.Join(sortFields, ", ")
}

// parseFieldValues 解析字段和值
func parseFieldValues(input string) (map[string]string, error) {
	if input == "" {
		return nil, nil
	}

	parts := strings.Split(input, ",")
	if len(parts)%2 != 0 {
		return nil, fmt.Errorf("参数格式错误：字段和值必须成对出现")
	}

	result := make(map[string]string)
	for i := 0; i < len(parts); i += 2 {
		field := strings.TrimSpace(parts[i])
		value := strings.TrimSpace(parts[i+1])

		if !SafeColumn(field) {
			return nil, fmt.Errorf("无效的字段名：%s", field)
		}

		result[field] = value
	}

	return result, nil
}

// parseInValues 解析IN查询的字段和值
func parseInValues(input string) (map[string][]string, error) {
	if input == "" {
		return nil, nil
	}

	parts := strings.Split(input, ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("参数格式错误：至少需要字段和一个值")
	}

	result := make(map[string][]string)
	currentField := ""
	values := make([]string, 0)

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if i == 0 {
			// 第一个值一定是字段名
			if !SafeColumn(part) {
				return nil, fmt.Errorf("无效的字段名：%s", part)
			}
			currentField = part
			continue
		}

		// 检查是否是新的字段名
		if SafeColumn(part) && i > 0 {
			// 保存之前的字段和值
			if currentField != "" && len(values) > 0 {
				result[currentField] = values
				values = make([]string, 0)
			}
			currentField = part
		} else {
			// 添加值
			values = append(values, part)
		}
	}

	// 保存最后一个字段和值
	if currentField != "" && len(values) > 0 {
		result[currentField] = values
	}

	return result, nil
}

// validateField 验证字段
func validateField(field, operator string, config *QueryConfig) error {
	// 检查字段是否在黑名单中
	if _, ok := config.Blacklist[field]; ok {
		return fmt.Errorf("字段 %s 被禁止查询", field)
	}

	// 如果配置了白名单，则检查字段是否在白名单中
	if len(config.Fields) > 0 {
		allowedOperators, ok := config.Fields[field]
		if !ok {
			return fmt.Errorf("不允许查询字段: %s", field)
		}

		// 检查操作符是否允许
		if !contains(allowedOperators, operator) {
			return fmt.Errorf("字段 %s 不支持 %s 操作符", field, operator)
		}
	}

	return nil
}

// validateAndBuildCondition 验证并构建查询条件
func validateAndBuildCondition(db *gorm.DB, input string, operator string, config *QueryConfig) error {
	var conditions map[string]string
	var err error

	if operator == "in" {
		inConditions, err := parseInValues(input)
		if err != nil {
			return err
		}
		for field, values := range inConditions {
			if err := validateField(field, operator, config); err != nil {
				return err
			}
			db = db.Where(field+" IN ?", values)
		}
		return nil
	}

	conditions, err = parseFieldValues(input)
	if err != nil {
		return err
	}

	for field, value := range conditions {
		if err := validateField(field, operator, config); err != nil {
			return err
		}

		switch operator {
		case "eq":
			db = db.Where(field+" = ?", value)
		case "like":
			db = db.Where(field+" LIKE ?", "%"+value+"%")
		case "gt":
			db = db.Where(field+" > ?", value)
		case "gte":
			db = db.Where(field+" >= ?", value)
		case "lt":
			db = db.Where(field+" < ?", value)
		case "lte":
			db = db.Where(field+" <= ?", value)
		}
	}

	return nil
}

// AutoPaginate 自动分页查询
func AutoPaginate[T any](
	ctx context.Context,
	db *gorm.DB,
	model interface{},
	data T,
	pageInfo *PageInfoReq,
	configs ...*QueryConfig,
) (*Paginated[T], error) {
	if pageInfo == nil {
		pageInfo = new(PageInfoReq)
	}

	// 构建查询条件
	if err := buildWhereConditions(db, pageInfo, configs...); err != nil {
		return nil, err
	}

	// 获取分页大小
	pageSize := pageInfo.GetLimit()
	offset := pageInfo.GetOffset()

	// 查询总数
	var totalCount int64
	if err := db.Debug().Model(model).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("分页查询统计总数失败: %w", err)
	}

	// 应用排序条件
	sortStr := pageInfo.GetSorts()
	if sortStr != "" {
		db = db.Order(sortStr)
	}

	// 查询当前页数据
	if err := db.Offset(offset).Limit(pageSize).Find(data).Error; err != nil {
		return nil, fmt.Errorf("分页查询数据失败: %w", err)
	}

	// 计算总页数
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize != 0 {
		totalPages++
	}

	return &Paginated[T]{
		Items:       data,
		CurrentPage: pageInfo.Page,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}

// buildWhereConditions 构建查询条件
func buildWhereConditions(db *gorm.DB, pageInfo *PageInfoReq, configs ...*QueryConfig) error {
	// 如果没有配置，直接构建查询条件
	if len(configs) == 0 {
		return buildWhereConditionsWithoutConfig(db, pageInfo)
	}

	// 合并所有配置
	config := mergeConfigs(configs...)

	// 验证并构建等于条件
	if err := validateAndBuildCondition(db, pageInfo.Eq, "eq", config); err != nil {
		return err
	}

	// 验证并构建模糊匹配条件
	if err := validateAndBuildCondition(db, pageInfo.Like, "like", config); err != nil {
		return err
	}

	// 验证并构建IN查询条件
	if err := validateAndBuildCondition(db, pageInfo.In, "in", config); err != nil {
		return err
	}

	// 验证并构建大于条件
	if err := validateAndBuildCondition(db, pageInfo.Gt, "gt", config); err != nil {
		return err
	}

	// 验证并构建大于等于条件
	if err := validateAndBuildCondition(db, pageInfo.Gte, "gte", config); err != nil {
		return err
	}

	// 验证并构建小于条件
	if err := validateAndBuildCondition(db, pageInfo.Lt, "lt", config); err != nil {
		return err
	}

	// 验证并构建小于等于条件
	if err := validateAndBuildCondition(db, pageInfo.Lte, "lte", config); err != nil {
		return err
	}

	return nil
}

// buildWhereConditionsWithoutConfig 无配置构建查询条件
func buildWhereConditionsWithoutConfig(db *gorm.DB, pageInfo *PageInfoReq) error {
	// 构建等于条件
	if pageInfo.Eq != "" {
		conditions, err := parseFieldValues(pageInfo.Eq)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" = ?", value)
			}
		}
	}

	// 构建模糊匹配条件
	if pageInfo.Like != "" {
		conditions, err := parseFieldValues(pageInfo.Like)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" LIKE ?", "%"+value+"%")
			}
		}
	}

	// 构建IN查询条件
	if pageInfo.In != "" {
		conditions, err := parseInValues(pageInfo.In)
		if err != nil {
			return err
		}
		for field, values := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" IN ?", values)
			}
		}
	}

	// 构建大于条件
	if pageInfo.Gt != "" {
		conditions, err := parseFieldValues(pageInfo.Gt)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" > ?", value)
			}
		}
	}

	// 构建大于等于条件
	if pageInfo.Gte != "" {
		conditions, err := parseFieldValues(pageInfo.Gte)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" >= ?", value)
			}
		}
	}

	// 构建小于条件
	if pageInfo.Lt != "" {
		conditions, err := parseFieldValues(pageInfo.Lt)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" < ?", value)
			}
		}
	}

	// 构建小于等于条件
	if pageInfo.Lte != "" {
		conditions, err := parseFieldValues(pageInfo.Lte)
		if err != nil {
			return err
		}
		for field, value := range conditions {
			if SafeColumn(field) {
				db = db.Where(field+" <= ?", value)
			}
		}
	}

	return nil
}

// mergeConfigs 合并多个配置
func mergeConfigs(configs ...*QueryConfig) *QueryConfig {
	merged := NewQueryConfig()

	for _, config := range configs {
		if config == nil {
			continue
		}

		// 合并白名单
		for field, operators := range config.Fields {
			if existing, ok := merged.Fields[field]; ok {
				existing = append(existing, operators...)
				existing = removeDuplicates(existing)
				merged.Fields[field] = existing
			} else {
				merged.Fields[field] = operators
			}
		}

		// 合并黑名单
		for field := range config.Blacklist {
			merged.Blacklist[field] = struct{}{}
		}
	}

	return merged
}

// removeDuplicates 去除切片中的重复元素
func removeDuplicates(slice []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

// contains 检查切片是否包含指定值
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
