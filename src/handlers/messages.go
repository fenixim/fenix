package handlers

import (
	"encoding/json"
	"fenix/src/models"
	"fenix/src/server"
	"log"
)

type MessageHandler struct {
	hub *server.ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("send_message", m.HandleSendMessage)
}

func (m *MessageHandler) HandleSendMessage(b []byte, client *server.Client) {
	var msg models.SendMessage
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	recv_msg := models.RecvMessage{
		T:       "recv_msg",
		Author:  client.Nick,
		Message: msg.Message,
	}

	client.OutgoingMessageQueue <- recv_msg
}

func NewMessageHandler(hub *server.ServerHub) *MessageHandler {
	m := MessageHandler{hub: hub}
	m.init()
	return &m
}
