package imserver

import (
	"fmt"
	"runtime/debug"

	"github.com/gorilla/websocket"
)

const (
	// 心跳超时时间，主要用于判断是否要断开链接
	heartbeatExpirationTime = 6 * 60
)

type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 用户连接
	Send          chan []byte     // 待发送的数据
	AppId         uint32          // 登录的平台Id app/web/ios
	UserId        string          // 用户Id，用户登录以后才有
	FirstTime     uint64          // 首次连接事件
	HeartbeatTime uint64          // 用户上次心跳时间
	LoginTime     uint64          // 登录时间 登录以后才有
}

// 初始化
func InitClient(addr string, socket *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		Addr:          addr,
		Socket:        socket,
		Send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}
	return client
}

func (c *Client) read() {
	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			fmt.Println("收到数据:", c.Addr, err)
			return
		}
		fmt.Println("收到数据:", string(message))
		// 收到数据后->数据处理
		HandleReceive(c, message)
	}
}

// 协程处理发送消息的队列
func (c *Client) write() {
	// 如果发送失败,说明连接出现问题->关闭连接
	defer func() {
		clientManager.Unregister <- c
		c.Socket.Close()
		fmt.Println("Client发送数据 defer", c)
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// 发送数据错误 关闭连接
				fmt.Println("Client发送数据 关闭连接", c.Addr, "ok", ok)
				return
			}
			c.Socket.WriteMessage(websocket.TextMessage, message)
		default:
			{
			}
		}

	}
}

// 发送消息
func (c *Client) SendMsg(msg []byte) {
	if c == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SendMsg stop:", r, string(debug.Stack()))
		}
	}()
	c.Send <- msg
}

// 心跳超时
func (c *Client) IsHeartbeatTimeout(currentTime uint64) (timeout bool) {
	if c.HeartbeatTime+heartbeatExpirationTime <= currentTime {
		timeout = true
	}
	return timeout
}

// 设置心跳
func (c *Client) SetHeardBeet(currentTime uint64) {
	c.HeartbeatTime = currentTime
}

// 登录->设置appId以及用户信息。
func (c *Client) Login(appId uint32, userId string, loginTime uint64) {
	c.AppId = appId
	c.UserId = userId
	c.LoginTime = loginTime
	// 登录成功=心跳一次
	c.SetHeardBeet(loginTime)
}

// 是否登录了
func (c *Client) IsLogin() (isLogin bool) {
	// 用户登录了
	if c.UserId != "" {
		isLogin = true
		return
	}
	return
}
