package db

import (
	"fmt"
	"github.com/yunhanshu-net/function-server/model"
	"time"

	"github.com/yunhanshu-net/function-server/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg config.DBConfig) error {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	}

	var err error
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取 *sql.DB 失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// 不自动迁移数据库表结构，假设表结构已存在
	err = DB.AutoMigrate(
		&model.Runner{},
		&model.FuncVersion{},
		&model.RunnerFunc{},
		&model.ServiceTree{},
		&model.RunnerVersion{},
		&model.FuncRunRecord{},
		&model.FunctionGen{},
	)
	if err != nil {
		return err
	}
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
