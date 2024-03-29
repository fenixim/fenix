package handlers

import (
	"encoding/json"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"time"
)

type MessageHandler struct {
	hub *server.ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("msg_send", m.HandleSendMessage)
	m.hub.RegisterHandler("msg_history", m.HandleMessageHistory)
}

func (m *MessageHandler) HandleSendMessage(b []byte, c *server.Client) {
	var msg websocket_models.MsgSend
	err := json.Unmarshal(b, &msg)
	if err != nil {
		utils.InfoLogger.Printf("error in decoding message json: %v", err)
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "JSONDecodeError"}
		return
	}

	if msg.Message == "" {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{
			Error:   "MessageEmpty",
			Message: "Cannot send an empty message!",
		}
		return
	}

	msg_broadcast := websocket_models.MsgBroadcast{
		Time: time.Now().UnixNano(),
		Author: websocket_models.Author{
			ID:       c.User.UserID.Hex(),
			Username: c.User.Username,
		},
		Message: msg.Message,
	}

	db_msg := database.Message{
		Content:   msg_broadcast.Message,
		Timestamp: msg_broadcast.Time,
		Author:    database.User{
			UserID:   c.User.UserID,
			Username: c.User.Username,
		},
	}

	err = m.hub.Database.InsertMessage(&db_msg)

	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
		return
	}

	msg_broadcast.MessageID = db_msg.MessageID.Hex()
	m.hub.Broadcast_payload <- msg_broadcast
}

func (m *MessageHandler) HandleMessageHistory(b []byte, c *server.Client) {
	hist := &websocket_models.MsgHistory{}
	err := json.Unmarshal(b, hist)
	if err != nil {
		utils.InfoLogger.Printf("error decoding message json, %v", err)
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "JSONDecodeError"}
		return
	}

	msgs, err := m.hub.Database.GetMessagesBetween(hist.From, hist.To, 50)
	if err != nil {
		utils.ErrorLogger.Printf("Error handling message history request: %q", err)
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
		return
	}

	hist.Messages = msgs

	c.OutgoingPayloadQueue <- hist
}

func NewMessageHandler(hub *server.ServerHub) *MessageHandler {
	m := MessageHandler{hub: hub}
	m.init()
	return &m
}
