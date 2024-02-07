package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"zinx-chat/protobuf/pb"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	// 消息ID
	Id uint32
	// 消息长度
	DataLen uint32
	// 消息内容
	Data []byte
}

// Pack 封包，将二进制数据转换为请求数据
func Pack(msg *Message) ([]byte, error) {
	// 初始化缓冲区，数据将会读到这里
	dataBuff := bytes.NewBuffer([]byte{})

	// 读取MsgId
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.Id); err != nil {
		return nil, err
	}

	// 读取DataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.DataLen); err != nil {
		return nil, err
	}

	// 读取数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.Data); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// Unpack 拆包，将请求数据转换为二进制数据
func Unpack(binaryData []byte) (*Message, error) {
	dataBuff := bytes.NewReader(binaryData)

	msg := &Message{}

	// 拆MsgId
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	// 拆DataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	return msg, nil
}

var localStorage sync.Map

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("Connect to 127.0.0.1:8999 error", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// 读
	go func() {
		defer wg.Done()
		for {
			// 收包、拆包
			headData := make([]byte, 8)
			_, err = io.ReadFull(conn, headData)
			if err != nil {
				fmt.Println("[Client]Recv head error", err)
				break
			}

			msgRecv, err := Unpack(headData)
			if err != nil {
				fmt.Println("[Client]Unpack head error", err)
				break
			}

			if msgRecv.DataLen > 0 {
				msgData := make([]byte, msgRecv.DataLen)
				_, err = io.ReadFull(conn, msgData)
				if err != nil {
					fmt.Println("[Client]Read msgData error", err)
					break
				}

				msgRecv.Data = msgData
			}

			switch msgRecv.Id {
			// 用户上线
			case 1:
				msg := &pb.SyncUser{}
				err := proto.Unmarshal(msgRecv.Data, msg)
				if err != nil {
					continue
				}

				uid, ok := localStorage.Load("uid")
				if !ok {
					localStorage.Store("uid", msg.Uid)
					localStorage.Store("username", msg.GetUsername())
					uid = msg.Uid
				}

				if msg.Uid != uid {
					fmt.Printf("用户[%s]进入聊天室\n", msg.GetUsername())
				}
			// 广播消息
			case 200:
				msg := &pb.Broadcast{}
				err := proto.Unmarshal(msgRecv.Data, msg)
				if err != nil {
					continue
				}

				switch msg.MsgType {
				// 下线
				case 2:
					fmt.Printf("用户[%s]离开聊天室\n", msg.GetUsername())
				// 全局聊天
				case 3:
					fmt.Printf("用户[%s]说: %s\n", msg.GetUsername(), msg.GetContent())
				}
			}
		}
	}()

	// 写
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			// 读取用户输入
			readStr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("输入的内容有误")
				continue
			}

			cutStr := strings.TrimSpace(readStr)
			if len(cutStr) == 0 {
				continue
			}

			// 组装成protobuf消息
			msg := &pb.Chat{
				Content: cutStr,
			}

			data, err := proto.Marshal(msg)
			if err != nil {
				fmt.Println("消息内容序列化失败")
				continue
			}

			// 组装成Message
			msgSend := &Message{
				Id:      uint32(2),
				DataLen: uint32(len(data)),
				Data:    data,
			}

			// 封包
			binaryData, err := Pack(msgSend)
			if err != nil {
				fmt.Println("[Client]Pack data error", err)
				continue
			}

			// 发包
			_, err = conn.Write(binaryData)
			if err != nil {
				fmt.Println("[Client]Send message error", err)
				break
			}
		}
	}()

	fmt.Println("成功连接到聊天室...")
	wg.Wait()
}
