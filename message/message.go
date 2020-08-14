package message

// 消息类型
type Type int8

const (
	// 心跳消息
	TypePing 	Type = 0
	// 业务消息
	TypeMessage Type = 1
)

// 出消息内容
type OutMessage struct {
	Code        int64       `json:"code"`
	Data        interface{} `json:"data"`
	Error       string      `json:"error"`
	MessageType Type        `json:"message_type"`
}

// 入消息内容
type InMessage struct {
	Service	string		`json:"service"`
	Data	interface{}	`json:"data"`
	Token	string		`json:"token"`
}

// 测试消息体
type Test struct {
	Id		int64	`json:"id"`
	Name	string	`json:"name"`
}
