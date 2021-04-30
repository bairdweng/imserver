# imserver
采用go构建简易即时通讯服务

快速上手

```shell
# 加载依赖
go mod tidy
# 运行起来
go run main.go
```

####  1. 配置篇

1. config/app.yaml中

   ```go
   app:
     logFile: log/gin.log
     httpPort: 8080
     webSocketPort: 8089
     rpcPort: 9001
     httpUrl: 127.0.0.1:8080
     webSocketUrl:  127.0.0.1:8089
   
   
   redis:
     addr: "localhost:6379"
     password: ""
     DB: 0
     poolSize: 30
     minIdleConns: 30
   ```

2. 初始化

   ```go
   func initConfig() {
   	viper.SetConfigName("config/app")
   	viper.AddConfigPath(".")
   	err := viper.ReadInConfig()
   	if err != nil {
   		println("读取配置出现异常")
   	}
   }
   ```

3. 读取配置信息的值

   ```go
   // 获取端口
   port := viper.GetString("app.webSocketPort")
   ```

#### 2. 启动一个websocket服务

```go
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
```
#### 3. 客户端与服务端建立连接并登录
1. 当客户端访问 **ws://localhost:8089/im ** 时，服务端将会开启类似 localhost:57759 以此建立一个连接，不同客户端产生不同的端口。

2. 建立成功之后，客户端发起登录的命令。

   ```js
   ws.send('{"seq":"' + sendId() + '","cmd":"login","data":{"userId":"' + person + '","appId":101}}');
   ```

3. 可以注意到 **cmd**，这是服务端约定好的命令 routers/websocket，在路由里面有相应的处理方法。

   ```go
   func InitWebSocket() {
   	// 注册websocket的登录命令
   	imserver.Register("login", imserver.HandleLogin)
   }
   ```
   
4. 心跳的处理，客户端不断发生心跳，服务端更新心跳的时间，当心跳超时，判断连接无效，关闭连接，节省资源。

   ```js
   // 客户端
   setInterval(heartbeat, 30 * 1000)
   ws.send('{"seq":"' + sendId() + '","cmd":"heartbeat","data":{}}');
   ```

   ```go
   // 注册心跳->会不断更新时间
   imserver.Register("heartbeat", imserver.HandleHearBeat)
   ```

   4.1 关闭连接

   ```js
   func ClearTimeoutConnections() {
   	currentTime := uint64(time.Now().Unix())
   	clients := clientManager.GetClients()
   	for client := range clients {
   		if client.IsHeartbeatTimeout(currentTime) {
   			fmt.Println("心跳时间超时 关闭连接", client.Addr, client.UserId, client.LoginTime, client.HeartbeatTime)
   
   			client.Socket.Close()
   		}
   	}
   }
   ```

#### 5. 关于协程的使用

1. 在imserver/manager.go中，可以看到有锁以及不同操作的通道。

   ```go
   // 客户连接管理
   type ClientManager struct {
   	Clients     map[*Client]bool   // 全部的连接
   	ClientsLock sync.RWMutex       // 读写锁
   	Users       map[string]*Client // 登录的用户
   	UserLock    sync.RWMutex       // 读写锁
   	Register    chan *Client       // 连接连接处理用于客户端的管理。
   	Unregister  chan *Client       // 断开连接处理程序
   }
   ```

2. 提取一个例子，在客户端建立连接的时候，client对象发送到通道中。Register是注册的通道

   2.1 发送数据

   ```go
   	// 加入manage统一管理
   	clientManager.Register <- client
   ```

   2.2 处理数据

   ```go
   func (manager *ClientManager) start() {
   	for {
   		select {
   		// 注册连接
   		case conn := <-manager.Register:
   			manager.EventRegister(conn)
   		//注销连接
   		case conn := <-manager.Unregister:
   			manager.EventUnregister(conn)
   		default:
   			{
   			}
   		}
   	}
   }
   ```

   2.3 加入到map中，注意加锁

   ```go
   // 用户建立连接事件
   func (manager *ClientManager) EventRegister(client *Client) {
   	manager.AddClient(client)
   	fmt.Println("EventRegister 用户建立连接", client.Addr)
   }
   
   // 添加客户端
   func (manager *ClientManager) AddClient(client *Client) {
   	// 协程安全
   	manager.ClientsLock.Lock()
   	defer manager.ClientsLock.Unlock()
   	manager.Clients[client] = true
   }
   ```

#### 6. 总结

1. 大概的流程是服务端建立websocket服务并确定端口，由于websocket是建立在http的基础上，所以会有个升级协议的过程。
2. 当客户端与服务端建立请求的时候，会产生对应的端口类似56378，同时会有一个socket对象，服务端可以建立一个连接对象，里面包含socket对象，用户id，过期时间等信息。
3. 每当建立连接，服务端将保存这个连接并统一管理。当需要发送信息或者A用户发送到B用户，服务端只需查询对于的用户，然后调用socket的sendMessage以及处理ReadMessage即可
4. 由于统一管理的所有的连接，同时客户端又有心跳的机制，那么服务端有个定时器，定期处理，筛选出已经过期的连接，关闭已节省服务端资源。

