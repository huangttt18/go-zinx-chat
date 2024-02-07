package core

import (
	"fmt"
	"sync"
)

const (
	// ChatTypeGlobal 聊天室类型: 全局
	ChatTypeGlobal int32 = iota
)

// ChatRoom 聊天室
type ChatRoom struct {
	// 房间ID
	RoomId int64
	// 类型
	ChatType int32
	// 在线用户列表
	Users map[int64]*User
	// 用户锁
	UserLock sync.RWMutex
}

// GlobalChatRoom 全局聊天室
var GlobalChatRoom *ChatRoom

func init() {
	GlobalChatRoom = &ChatRoom{
		Users:    make(map[int64]*User),
		ChatType: ChatTypeGlobal,
	}
}

// NewChatRoom 新建一个聊天室，可以指定聊天室类型
func NewChatRoom(chatType int32) *ChatRoom {
	chatRoom := &ChatRoom{
		Users: make(map[int64]*User),
	}

	return chatRoom
}

// Add 新增用户
func (cr *ChatRoom) Add(u *User) {
	cr.UserLock.Lock()
	defer cr.UserLock.Unlock()

	if _, ok := cr.Users[u.Uid]; ok {
		fmt.Printf("用户[%d]已在线\n", u.Uid)
		return
	}

	cr.Users[u.Uid] = u
}

// Remove 删除用户
func (cr *ChatRoom) Remove(uid int64) {
	cr.UserLock.Lock()
	defer cr.UserLock.Unlock()

	_, ok := cr.Users[uid]
	if !ok {
		fmt.Printf("用户[%d]不在线\n", uid)
		return
	}

	delete(cr.Users, uid)
}

// Get 获取单个用户
func (cr *ChatRoom) Get(uid int64) *User {
	cr.UserLock.RLock()
	defer cr.UserLock.RUnlock()

	if user, ok := cr.Users[uid]; ok {
		return user
	}

	return nil
}

// ListUids 获取全部在线用户ID列表
func (cr *ChatRoom) ListUids() []int64 {
	cr.UserLock.RLock()
	defer cr.UserLock.RUnlock()

	uids := make([]int64, 0)
	for _, user := range cr.Users {
		uids = append(uids, user.Uid)
	}

	return uids
}

// ListUsers 获取全部在线用户列表
func (cr *ChatRoom) ListUsers() []*User {
	cr.UserLock.RLock()
	defer cr.UserLock.RUnlock()

	users := make([]*User, 0)
	for _, user := range cr.Users {
		users = append(users, user)
	}

	return users
}
