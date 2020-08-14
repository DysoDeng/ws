package connection

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
	"ws/message"
)

// 连接信息
type Connection struct {
	// websocket 连接
	wsConn 		*websocket.Conn
	// 入消息内容
	inMessage	chan message.InMessage
	// 出消息内容
	outMessage	chan message.OutMessage
	// 关闭连接
	closeChan	chan byte
	// 连接锁
	mutex		sync.Mutex
	// 连接是否被关闭
	isClosed	bool
	// 用户ID
	userId		int64
	// 加入的群组
	groups		map[string]bool
}

// 群信息
type Group struct {
	Clients map[int64]*Connection
}

var (
	Groups = make(map[string]*Group)
	groupLock sync.Mutex
)

// 创建连接
func NewConnection(wsConn *websocket.Conn) (conn *Connection, err error) {

	conn = &Connection{
		wsConn:     wsConn,
		inMessage:  make(chan message.InMessage, 1000),
		outMessage: make(chan message.OutMessage, 1000),
		closeChan:  make(chan byte, 1),
		isClosed:   false,
		groups:		make(map[string]bool),
	}

	// 启动心跳协程
	go conn.ping()

	// 启动读协程
	go conn.readLoop()

	// 启动写协程
	go conn.writeLoop()

	return
}

// 消息群发
func WriteMessageAll(groupId string, outMessage message.OutMessage) (err error) {
	data, err := json.Marshal(outMessage)
	if err == nil {
		for c := range Groups[groupId].Clients {
			if err = Groups[groupId].Clients[c].wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
				Groups[groupId].Clients[c].Close()
			}
		}
	}

	return
}

// 加入群组
func (conn *Connection) AddGroup(groupId string, userId int64) {
	groupLock.Lock() // 互斥锁
	if _, ok := Groups[groupId]; ok {
		Groups[groupId].Clients[userId] = conn
	} else {
		Groups[groupId] = &Group{
			Clients:   make(map[int64]*Connection),
		}
		Groups[groupId].Clients[userId] = conn
	}
	conn.groups[groupId] = true
	conn.userId = userId
	groupLock.Unlock()
}

// 退出群组
func (conn *Connection) ExitGroup(groupId string) {
	if _, ok := conn.groups[groupId]; ok {
		if _, o := Groups[groupId].Clients[conn.userId]; o {
			delete(Groups[groupId].Clients, conn.userId)
		}
	}
}

// 读取消息
func (conn *Connection) ReadMessage() (inMessage message.InMessage, err error) {

	select {
	case inMessage = <-conn.inMessage:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

// 发送消息
func (conn *Connection) WriteMessage(outMessage message.OutMessage) (err error) {

	select {
	case conn.outMessage <- outMessage:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

// 关闭连接
func (conn *Connection) Close() {
	// 线程安全的Close，可以并发多次调用也叫做可重入的Close
	_ = conn.wsConn.Close()
	conn.mutex.Lock()
	if !conn.isClosed {
		// 退出群组
		for g := range conn.groups {
			if _,ok := Groups[g]; ok {
				if _,ok := Groups[g].Clients[conn.userId]; ok {
					delete(Groups[g].Clients, conn.userId)
				}
			}
		}
		// 关闭chan,但是chan只能关闭一次
		close(conn.closeChan)
		conn.isClosed = true
		log.Println("connection is closed!!!")
	}
	conn.mutex.Unlock()
}

// 读取ws消息
func (conn *Connection) readLoop() {
	var (
		data []byte
		err error
	)

	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}

		// 处理消息
		var inMessage message.InMessage
		err = json.Unmarshal(data, &inMessage)
		if err != nil {
			_ = conn.WriteMessage(message.OutMessage{
				Code:  500,
				Data:  nil,
				Error: "消息错误",
			})
		} else {
			select {
			case conn.inMessage <- inMessage:
			case <- conn.closeChan:
				goto ERR
			}
		}
	}

	ERR:
		conn.Close()
}

// 发送ws消息
func (conn *Connection) writeLoop() {
	var (
		data       []byte
		outMessage message.OutMessage
		err        error
	)

	for {
		select {
		case outMessage = <- conn.outMessage: // 消息一对一发
			data, err = json.Marshal(outMessage)
			if err != nil {
				goto ERR
			}
			if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
				goto ERR
			}
		case <- conn.closeChan: // 连接关闭
			goto ERR
		}
	}

	ERR:
		conn.Close()
}

// 发送心跳
func (conn *Connection) ping() {
	var err error
	for {
		// 每20秒发送一次心跳
		if err = conn.WriteMessage(message.OutMessage{
			Code:        0,
			Data:        "ping",
			Error:       "",
			MessageType: message.TypePing,
		}); err != nil {
			return
		}
		time.Sleep(20 * time.Second)
	}
}