package main

import (
	"com.bwtg.im/imserver"
	"com.bwtg.im/task"
	"github.com/spf13/viper"
)

func main() {
	initConfig()
	// 定时任务
	task.Start()
	// 启动websocket服务
	imserver.Start()
}

// 初始化配置信息
func initConfig() {
	viper.SetConfigName("config/app")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		println("读取配置出现异常")
	}
}
