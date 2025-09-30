package model

import "time"

// CdnFile CDN文件分享表
type CdnFile struct {
	Base
	ShareKey      string     `json:"share_key" gorm:"column:share_key;type:varchar(255);uniqueIndex;comment:分享密钥"`
	FileName      string     `json:"file_name" gorm:"column:file_name;comment:原始文件名"`
	FileSize      int64      `json:"file_size" gorm:"column:file_size;comment:文件大小(字节)"`
	FileType      string     `json:"file_type" gorm:"column:file_type;comment:文件类型"`
	FileURL       string     `json:"file_url" gorm:"column:file_url;comment:文件下载URL"`
	Password      string     `json:"password" gorm:"column:password;comment:访问密码(加密存储)"`
	MaxDownloads  int        `json:"max_downloads" gorm:"column:max_downloads;comment:最大下载次数"`
	DownloadCount int        `json:"download_count" gorm:"column:download_count;comment:已下载次数"`
	ExpiresAt     *time.Time `json:"expires_at" gorm:"column:expires_at;comment:过期时间"`
	Description   string     `json:"description" gorm:"column:description;comment:文件描述"`
	User          string     `json:"user" gorm:"column:user;comment:上传用户"`
}

// TableName 表名
func (CdnFile) TableName() string {
	return "cdn_file"
}
