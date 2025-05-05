package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用程序的配置结构
type Config struct {
	ServerConfig  ServerConfig  `json:"server"`
	DBConfig      DBConfig      `json:"database"`
	LogConfig     LogConfig     `json:"log"`
	RuncherConfig RuncherConfig `json:"runcher"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Mode         string `json:"mode"` // debug, release
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Type         string `json:"type"` // mysql, postgres, sqlite
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxLifetime  int    `json:"max_lifetime"` // seconds
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"` // debug, info, warn, error
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // days
	Compress   bool   `json:"compress"`
}

// RuncherConfig Runcher服务配置
type RuncherConfig struct {
	NatsURL string `json:"nats_url"` // NATS服务器URL
	Timeout int    `json:"timeout"`  // 请求超时时间（秒）
}

var config Config

// Init 初始化配置
func Init() error {
	// 先尝试从环境变量获取配置文件路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 默认配置文件路径
		configPath = "configs/config.json"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果不存在，创建默认配置
		defaultConfig := Config{
			ServerConfig: ServerConfig{
				Port:         8080,
				Mode:         "debug",
				ReadTimeout:  60,
				WriteTimeout: 60,
			},
			DBConfig: DBConfig{
				Type:         "mysql",
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Password:     "password",
				Name:         "yunhanshu",
				MaxIdleConns: 10,
				MaxOpenConns: 100,
				MaxLifetime:  3600,
			},
			LogConfig: LogConfig{
				Level:      "debug",
				Filename:   "logs/app.log",
				MaxSize:    100,
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
			},
			RuncherConfig: RuncherConfig{
				NatsURL: "nats://localhost:4222",
				Timeout: 20,
			},
		}

		// 创建配置文件
		file, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("创建配置文件失败: %w", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(defaultConfig); err != nil {
			return fmt.Errorf("写入默认配置失败: %w", err)
		}

		config = defaultConfig
		return nil
	}

	// 读取配置文件
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// Get 获取配置
func Get() *Config {
	return &config
}
