package service

import (
	"github.com/goinggo/mapstructure"
	"log"
	"ws/connection"
	"ws/message"
)

type User struct {
	Id			int64	`json:"id"`
	Username 	string	`json:"username"`
}

// 服务分发
func Service(conn *connection.Connection, inMessage message.InMessage) (err error) {
	// TODO Token鉴权

	log.Println(inMessage)

	// 服务处理
	switch inMessage.Service {
	case "Test.message":
		switch inMessage.Data.(type) {
		case string:
			if err = conn.WriteMessage(message.OutMessage{
				Code:        0,
				Data:        "这是字符串消息",
				Error:       "",
				MessageType: message.TypeMessage,
			}); err != nil {
				return
			}
		case map[string]interface{}:
			var data message.Test
			if err := mapstructure.Decode(inMessage.Data, &data); err != nil {
				if err = conn.WriteMessage(message.OutMessage{
					Code:        0,
					Data:        "消息错误",
					Error:       "",
					MessageType: message.TypeMessage,
				}); err != nil {

				}
			} else {
				if err = conn.WriteMessage(message.OutMessage{
					Code:        0,
					Data:        message.Test{
						Id:   data.Id,
						Name: data.Name,
					},
					Error:       "",
					MessageType: message.TypeMessage,
				}); err != nil {

				}
			}
		}
		break
	case "in_group": // 进入群
		switch inMessage.Data.(type) {
		case map[string]interface{}:
			var user User
			if err := mapstructure.Decode(inMessage.Data, &user); err == nil {
				if user.Id > 0 {
					conn.AddGroup("123", user.Id)

					if err = connection.WriteMessageAll("123", message.OutMessage{
						Code:        0,
						Data:        "用户["+user.Username+"]进入房间",
						Error:       "",
						MessageType: message.TypeMessage,
					}); err != nil {

					}
				}
			}
		}

		break
	case "exit_group": // 退出群
		conn.ExitGroup("123")
		if err = connection.WriteMessageAll("123", message.OutMessage{
			Code:        0,
			Data:        "我退出群啦",
			Error:       "",
			MessageType: message.TypeMessage,
		}); err != nil {
			return
		}
	case "message":
		if err = connection.WriteMessageAll("123", message.OutMessage{
			Code:        0,
			Data:        inMessage.Data,
			Error:       "",
			MessageType: message.TypeMessage,
		}); err != nil {
			return
		}
		break
	}

	return
}
