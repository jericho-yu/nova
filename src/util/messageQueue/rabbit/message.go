package rabbit

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jericho-yu/nova/src/util/str"

	"github.com/google/uuid"
)

type (
	Message struct {
		err     error
		MsgId   string `json:"msgId" yaml:"msgId"`
		Content []byte `json:"content" yaml:"content"`
	}
)

var (
	MessageApp Message
)

// New 实例化：创建消息
func (*Message) New(message []byte) *Message {
	return &Message{MsgId: uuid.Must(uuid.NewV6()).String(), Content: message}
}

// Parse 实例化：通过原始消息解析
func (*Message) Parse(prototypeMessage []byte) *Message {
	prototypeMessages := strings.Split(string(prototypeMessage), ":")

	return &Message{
		MsgId:   prototypeMessages[0],
		Content: []byte(prototypeMessages[1]),
	}
}

// ToJson 序列化：json
func (my *Message) ToJson() []byte {
	var (
		err     error
		content []byte
	)

	content, err = json.Marshal(my.Content)
	if err != nil {
		my.err = fmt.Errorf("序列化错误：%w", err)
		return nil
	}

	return str.BufferApp.NewByBytes([]byte(my.MsgId)).String(string(content)).ToBytes()
}
