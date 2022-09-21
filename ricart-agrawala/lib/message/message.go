package message

import (
	"bytes"
	"encoding/gob"
	"log"
)

type MessageType string

const (
	Request MessageType = "REQUEST"
	Reply   MessageType = "REPLY"
	Acquire MessageType = "ACQUIRE"
)

type Message struct {
	From      int
	Timestamp int
	Type      MessageType
}

func NewMessage(fromId int, timestamp int, msgType MessageType) *Message {
	return &Message{
		From:      fromId,
		Timestamp: timestamp,
		Type:      msgType,
	}
}

func (m *Message) EncodeToBytes() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func DecodeToMessage(s []byte) *Message {
	p := &Message{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
