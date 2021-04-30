package routers

import "com.bwtg.im/imserver"

func InitWebSocket() {
	// 注册websocket的登录命令
	imserver.Register("login", imserver.HandleLogin)
	// 注册心跳->会不断更新时间
	imserver.Register("heartbeat", imserver.HandleHearBeat)

}
