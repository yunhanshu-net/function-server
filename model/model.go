// model.go
// 模型包的主入口文件
// 导入所有子包，方便应用层代码使用
package model

import (
	// 导入所有子包，确保它们被初始化
	_ "github.com/yunhanshu-net/api-server/model/base"
	_ "github.com/yunhanshu-net/api-server/model/directory"
	_ "github.com/yunhanshu-net/api-server/model/runcher"
	_ "github.com/yunhanshu-net/api-server/model/runner"
	_ "github.com/yunhanshu-net/api-server/model/version"
)

// 这里提供一些跨子包的公共函数或常量
// 设计合理的话，大多数逻辑应该放在各自的子包中
// 这个文件主要起到导入整合的作用
