package imserver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"com.bwtg.im/common"
	"com.bwtg.im/models"
)

type DisposeFunc func(client *Client, seq string, message []byte) (code uint32, msg string, data interface{})

var (
	handlers        = make(map[string]DisposeFunc)
	handlersRWMutex sync.RWMutex
)

func getHandlers(key string) (value DisposeFunc, ok bool) {
	handlersRWMutex.RLock()
	defer handlersRWMutex.RUnlock()
	value, ok = handlers[key]
	return
}

// 处理登录
func HandleLogin(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	// 初始化请求对象
	request := &models.Login{}
	// json绑定数据
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		fmt.Println("用户登录 解析数据失败", seq, err)
		return
	}
	// 在这里做用户校验
	if request.UserId == "" || len(request.UserId) >= 20 {
		code = common.UnauthorizedUserId
		fmt.Println("用户登录 非法的用户", seq, request.UserId)
		return
	}
	// 用户已登录
	if client.IsLogin() {
		fmt.Println("用户登录 用户已经登录", client.AppId, client.UserId, seq)
		code = common.OperationFailure
		return
	}
	currentTime := uint64(time.Now().Unix())
	client.Login(request.AppId, request.UserId, currentTime)

	return
}

// 处理心跳
func HandleHearBeat(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	currentTime := uint64(time.Now().Unix())
	request := &models.HeartBeat{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		fmt.Println("心跳接口 解析数据失败", seq, err)
		return
	}
	if !client.IsLogin() {
		fmt.Println("心跳接口 用户未登录", client.AppId, client.UserId, seq)
		code = common.NotLoggedIn
		return
	}
	// 更新心跳时间
	client.SetHeardBeet(currentTime)
	return
}

// 服务端收到客户端的数据，开始处理
func HandleReceive(client *Client, message []byte) {
	// 解析数据
	request := &models.Request{}
	err := json.Unmarshal(message, request)
	if err != nil {
		fmt.Println("处理数据 json Unmarshal", err)
		client.SendMsg([]byte("数据不合法"))
		return
	}
	// 解析失败
	requestData, err := json.Marshal(request.Data)
	if err != nil {
		fmt.Println("处理数据 json Marshal", err)
		client.SendMsg([]byte("处理数据失败"))
		return
	}
	seq := request.Seq
	cmd := request.Cmd
	var (
		code uint32
		msg  string
		data interface{}
	)
	// 判断路由是否存在
	if value, ok := getHandlers(cmd); ok {
		code, msg, data = value(client, seq, requestData)
	} else {
		code = common.RoutingNotExist
		fmt.Println("处理数据 路由不存在", client.Addr, "cmd", cmd)
	}
	// 包装好数据->发送
	msg = common.GetErrorMessage(code, msg)
	responseHead := models.NewResponseHead(seq, cmd, code, msg, data)
	headByte, err := json.Marshal(responseHead)
	if err != nil {
		fmt.Println("处理数据 json Marshal", err)

		return
	}
	client.SendMsg(headByte)
	fmt.Println("acc_response send", client.Addr, client.AppId, client.UserId, "cmd", cmd, "code", code)
}
