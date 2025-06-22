package service

import (
	"context"
	"time"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/yunhanshu-net/function-server/pkg/config"
)

// OSSService OSS服务接口
type OSSService interface {
	// GetUploadToken 获取上传token
	GetUploadToken(ctx context.Context) (string, error)
}

// QiNiuOSSService 七牛云OSS服务实现
type QiNiuOSSService struct {
	config *config.QiNiuConfig
	mac    *qbox.Mac
}

// NewQiNiuOSSService 创建七牛云OSS服务
func NewQiNiuOSSService(cfg *config.QiNiuConfig) *QiNiuOSSService {
	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)
	return &QiNiuOSSService{
		config: cfg,
		mac:    mac,
	}
}

// GetUploadToken 获取上传token
func (s *QiNiuOSSService) GetUploadToken(ctx context.Context) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope:   s.config.Bucket,
		Expires: uint64(time.Now().Add(time.Hour).Unix()), // 1小时过期
	}

	upToken := putPolicy.UploadToken(s.mac)
	return upToken, nil
}

// GetConfig 获取七牛云配置信息
func (s *QiNiuOSSService) GetConfig() *config.QiNiuConfig {
	return s.config
}
 