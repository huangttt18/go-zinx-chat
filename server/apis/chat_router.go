package apis

import (
	"fmt"
	"zinx-chat/protobuf/pb"
	"zinx-chat/server/core"
	"zinx-demo/ziface"
	"zinx-demo/znet"

	"google.golang.org/protobuf/proto"
)

// ChatApi 聊天路由
type ChatApi struct {
	znet.BaseRouter
}

func (cr *ChatApi) Handle(request ziface.IRequest) {
	// 聊天内容反序列化
	msg := &pb.Chat{}
	err := proto.Unmarshal(request.GetData(), msg)
	if err != nil {
		fmt.Println("Proto unmarshal失败", err)
		return
	}

	// 获取到发送消息的用户id
	uid, err := request.GetConnection().GetProperty("uid")
	if err != nil {
		fmt.Println("获取用户uid失败")
		return
	}

	// 根据uid获取用户，并发送Chat消息
	user := core.GlobalChatRoom.Get(uid.(int64))
	user.Chat(msg.Content)
}
