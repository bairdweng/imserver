package imserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var (
	clientManager = NewClientManager()
)

func Start() {
	// websocket的端口
	port := viper.GetString("app.webSocketPort")
	// 处理协议
	http.HandleFunc("/im", webSocketConnect)
	fmt.Println("websocket已经启动:", "ws://127.0.0.1:"+port)
	// 开启协程处理事件，需要放在http端口之前。
	go clientManager.start()
	// 监听端口
	error := http.ListenAndServe(":"+port, nil)
	if error != nil {
		fmt.Println("启动服务异常:", error.Error())
	}
}

func webSocketConnect(w http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		fmt.Println("升级协议", "ua:", r.Header["User-Agent"], "referer:", r.Header["Referer"])
		return true
	}}).Upgrade(w, req, nil)
	if err != nil {
		fmt.Println("startWebSocket error:", err.Error())
		http.NotFound(w, req)
		return
	}
	// 建立连接之后会产生新的端口 57759
	address := conn.RemoteAddr().String()
	// 当前时间
	currentTime := uint64(time.Now().Unix())
	// 初始化客户端
	client := InitClient(address, conn, currentTime)
	// 开启协程读取数据
	go client.read()
	// 开启协程写入数据
	go client.write()
	// 加入manage统一管理
	clientManager.Register <- client
}
