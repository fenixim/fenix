package server

import (
	"encoding/json"
	"fenix/src/models"
	"log"
	"time"
)

type MessageHandler struct {
	hub *ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("send_message", m.HandleSendMessage)
}

func (m *MessageHandler) HandleSendMessage(b []byte, client *Client) {
	var msg models.SendMessage
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	recv_msg := models.RecvMessage{
		T:       "recv_msg",
		Time:    time.Now().Unix(),
		Author:  client.Nick,
		Message: msg.Message,
	}

	client.OutgoingPayloadQueue <- recv_msg
}

func NewMessageHandler(hub *ServerHub) *MessageHandler {
	m := MessageHandler{hub: hub}
	m.init()
	return &m
}
