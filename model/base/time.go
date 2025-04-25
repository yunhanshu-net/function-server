// time.go
// 提供自定义时间类型，处理时间的序列化和数据库交互
package base

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Time 自定义时间类型
// 实现了JSON序列化/反序列化和数据库扫描/保存
type Time time.Time

// 时间格式常量
const ctLayout = "2006-01-02 15:04:05"

// UnmarshalJSON 实现JSON反序列化
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(ctLayout, s)
	*t = Time(nt)
	return
}

// GetUnix 获取Unix时间戳
func (t Time) GetUnix() int64 {
	return time.Time(t).Unix()
}

// MarshalJSON 实现JSON序列化
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// String 格式化为字符串
func (t Time) String() string {
	return fmt.Sprintf("%q", time.Time(t).Format(ctLayout))
}

// Scan 从数据库扫描
func (date *Time) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*date = Time(nullTime.Time)
	return
}

// Value 转换为数据库值
func (date Time) Value() (driver.Value, error) {
	ti := time.Time(date)
	y, m, d := ti.Date()
	h := ti.Hour()
	minute := ti.Minute()
	s := ti.Second()
	return time.Date(y, m, d, h, minute, s, 0, time.Time(date).Location()), nil
}
