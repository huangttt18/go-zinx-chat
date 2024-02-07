package core

import (
	"fmt"
	"sync/atomic"
	"zinx-chat/protobuf/pb"
	"zinx-demo/ziface"

	"google.golang.org/protobuf/proto"
)

// User 用户
type User struct {
	// 用户ID
	Uid int64
	// 用户名
	Username string
	// 在线状态
	Status int8
	// 当前连接
	Conn ziface.IConnection
}

const (
	// StatusOffline 下线
	StatusOffline = 0
	// StatusOnline 上线
	StatusOnline
)

// UidGen 全局uid生成器
var UidGen int64 = 0

func NewUser(username string, conn ziface.IConnection) *User {
	user := &User{
		Username: username,
		Status:   StatusOnline,
		Conn:     conn,
	}

	// 生成用户id
	atomic.AddInt64(&UidGen, 1)
	user.Uid = UidGen

	return user
}

// Online 上线之后广播消息
func (u *User) Online() {
	msg := &pb.SyncUser{
		Uid:      u.Uid,
		Username: u.Username,
	}

	users := GlobalChatRoom.ListUsers()
	for _, user := range users {
		user.SendMsg(1, msg)
	}
}

// Offline 下线之后广播消息
func (u *User) Offline() {
	msg := &pb.Broadcast{
		Sender:   u.Uid,
		Username: u.Username,
		MsgType:  2,
	}

	users := GlobalChatRoom.ListUsers()
	for _, user := range users {
		user.SendMsg(200, msg)
	}
}

func (u *User) Chat(content string) {
	// 构建广播消息
	msg := &pb.Broadcast{
		Sender:   u.Uid,
		Username: u.Username,
		MsgType:  3,
		Data: &pb.Broadcast_Content{
			Content: content,
		},
	}

	// 获取当前在线用户列表，并广播聊天内容
	users := GlobalChatRoom.ListUsers()
	for _, user := range users {
		user.SendMsg(200, msg)
	}
}

// SendMsg 发送消息给客户端
func (u *User) SendMsg(msgId uint32, data proto.Message) {
	if u.Conn == nil {
		fmt.Printf("用户[%d]连接已断开，发送消息失败\n", u.Uid)
		return
	}

	// 序列化protobuf
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("反序列化消息失败", err)
		return
	}

	// 发送给客户端
	if err := u.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Printf("用户[%d]消息发送失败\n", u.Uid)
		return
	}
}
