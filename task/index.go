package task

import (
	"fmt"
	"time"

	"com.bwtg.im/imserver"
)

func Start() {
	Timer(3*time.Second, 5*time.Second, log, "", nil, nil)
}

func log(interface{}) (result bool) {
	fmt.Println("当前在线数:", imserver.GetClients())
	return true
}
