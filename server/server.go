package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	rand2 "math/rand"
	"strconv"
	"zinx-chat/server/apis"
	"zinx-chat/server/core"
	"zinx-demo/ziface"
	"zinx-demo/znet"
)

func generateRandomUsername(len int) string {
	randomBytes := make([]byte, len)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("生成用户名异常")
		return "Chat_" + strconv.Itoa(rand2.Intn(10))
	}
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return "Chat_" + randomString[:len]
}

// OnConnStart 建立连接后需要做的操作
func OnConnStart(conn ziface.IConnection) {
	// 注册用户
	user := core.NewUser(generateRandomUsername(10), conn)
	// 将用户添加到聊天室中
	core.GlobalChatRoom.Add(user)
	// 将当前用户ID绑定到连接中
	conn.SetProperty("uid", user.Uid)
	// 广播上线消息
	user.Online()
}

// OnConnStop 断开连接后需要做的操作
func OnConnStop(conn ziface.IConnection) {
	// 获取到当前用户的uid
	uid, err := conn.GetProperty("uid")
	if err != nil {
		fmt.Printf("用户[%d]不存在", uid.(int64))
		return
	}

	// 获取到当前用户
	user := core.GlobalChatRoom.Get(uid.(int64))
	// 广播下线消息
	user.Offline()
	// 将当前用户从聊天室移除
	core.GlobalChatRoom.Remove(uid.(int64))
}

func main() {
	server := znet.NewServer("Zinx Chat Server V0.1")
	// 注册Hook函数
	server.SetOnConnStart(OnConnStart)
	server.SetOnConnStop(OnConnStop)
	// 注册路由
	server.AddRouter(2, &apis.ChatApi{})
	server.Serve()
}
