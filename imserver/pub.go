package imserver

// 获取客户端总数
func GetClients() int {
	return len(clientManager.Clients)
}

// 注册登录的操作
func Register(key string, value DisposeFunc) {
	handlersRWMutex.Lock()
	defer handlersRWMutex.Unlock()
	handlers[key] = value
}
