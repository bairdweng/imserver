package imserver

import (
	"fmt"
	"sync"
	"time"
)

// 客户连接管理
type ClientManager struct {
	Clients     map[*Client]bool   // 全部的连接
	ClientsLock sync.RWMutex       // 读写锁
	Users       map[string]*Client // 登录的用户
	UserLock    sync.RWMutex       // 读写锁
	Register    chan *Client       // 连接连接处理用于客户端的管理。
	Unregister  chan *Client       // 断开连接处理程序
}

func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:  make(map[*Client]bool),
		Users:    make(map[string]*Client),
		Register: make(chan *Client, 1000),
	}
	return clientManager
}

//  结构体的方法
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

// 释放连接
func (manager *ClientManager) EventUnregister(client *Client) {
	manager.DeleteClient(client)
}

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

// 删除客户端
func (manager *ClientManager) DeleteClient(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()
	// 判断是否存在->删除
	if manager.Clients[client] {
		delete(manager.Clients, client)
	}
}

// 定时清理超时连接
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

func (manager *ClientManager) GetClients() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	manager.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value
		return true
	})

	return
}
func (manager *ClientManager) ClientsRange(f func(client *Client, value bool) (result bool)) {
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()
	for key, value := range manager.Clients {
		result := f(key, value)
		if !result {
			return
		}
	}
	return
}
