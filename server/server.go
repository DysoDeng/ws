package server

import (
	"github.com/gorilla/websocket"
	"net/http"
	"ws/connection"
	"ws/message"
	"ws/service"
)

var (
	upgrade = websocket.Upgrader{
		// 读取存储空间大小
		ReadBufferSize:1024,

		// 写入存储空间大小
		WriteBufferSize:1024,

		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// websocket 请求
func WsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn    *websocket.Conn
		err       error
		conn      *connection.Connection
		inMessage message.InMessage
	)

	// 升级http请求为websocket请求
	if wsConn, err = upgrade.Upgrade(w, r, nil); err != nil {
		return
	}

	// 建立新的ws连接
	if conn, err = connection.NewConnection(wsConn); err != nil {
		goto ErrWs
	}

	// 处理消息
	for {

		if inMessage, err = conn.ReadMessage(); err != nil {
			goto Err
		}

		// 服务分发
		err = service.Service(conn, inMessage)
		if err != nil {
			goto Err
		}
	}

	ErrWs:
		wsConn.Close()
	Err:
		conn.Close()
}
