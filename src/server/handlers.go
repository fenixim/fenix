package server

import (
	"encoding/json"
	"fenix/src/database"
	"fenix/src/websocket_models"
	"log"
	"time"
)


type MessageHandler struct {
	hub *ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("msg_send", m.HandleSendMessage)
	m.hub.RegisterHandler("msg_history", m.HandleMessageHistory)
}

func (m *MessageHandler) HandleSendMessage(b []byte, client *Client) {
	var msg websocket_models.SendMessage
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	if msg.Message == "" {
		client.OutgoingPayloadQueue <- websocket_models.GenericError{
			Error: "message_empty",
			Message: "Cannot send an empty message!",
		}
		return
	}

	msg_broadcast := websocket_models.BroadcastMessage{
		Time: time.Now().Unix(),
		Author: websocket_models.Author{
			ID:       client.User.UserID.Hex(),
			Username: client.User.Username,
		},
		Message: msg.Message,
	}

	db_msg := database.Message{
		Content:   msg_broadcast.Message,
		Timestamp: msg_broadcast.Time,
		Author:    msg_broadcast.Author.ID,
	}

	err = m.hub.Database.InsertMessage(&db_msg)
	if err != nil {
		client.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
	}

	msg_broadcast.MessageID = db_msg.MessageID.Hex()
	m.hub.broadcast_payload <- msg_broadcast
}

func (m *MessageHandler) HandleMessageHistory(b []byte, client *Client) {
	hist := &websocket_models.MessageHistory{}
	json.Unmarshal(b, hist)
	if hist.From == 0 || hist.To == 0 {
		client.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "BadFormat", Message: "MessageHistory needs From and To fields"}
		return
	}

	msgs, err := m.hub.Database.GetMessagesBetween(hist.From, hist.To, 50)
	if err != nil {
		log.Printf("error in HandleMessageHistory, %v", err)
		return
	}
	hist.Messages = msgs

	client.OutgoingPayloadQueue <- hist
}


func NewMessageHandler(hub *ServerHub) *MessageHandler {
	m := MessageHandler{hub: hub}
	m.init()
	return &m
}


type IdentificationHandler struct {
	hub *ServerHub
}

func (i *IdentificationHandler) init() {
	i.hub.RegisterHandler("whoami", i.HandleWhoAmI)
}

func (i *IdentificationHandler) HandleWhoAmI(_ []byte, c *Client) {
	c.OutgoingPayloadQueue <- websocket_models.WhoAmI{
		ID:       c.User.UserID.Hex(),
		Username: c.User.Username,
	}
}

func NewIdentificationHandler(hub *ServerHub) {
	i := IdentificationHandler{hub: hub}
	i.init()
}
