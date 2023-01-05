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
	var msg websocket_models.MsgSend
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	if msg.Message == "" {
		client.OutgoingPayloadQueue <- websocket_models.GenericError{
			Error:   "message_empty",
			Message: "Cannot send an empty message!",
		}
		return
	}

	msg_broadcast := websocket_models.MsgBroadcast{
		Time: time.Now().UnixNano(),
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
		return
	}

	msg_broadcast.MessageID = db_msg.MessageID.Hex()
	m.hub.broadcast_payload <- msg_broadcast
}

func (m *MessageHandler) HandleMessageHistory(b []byte, client *Client) {
	hist := &websocket_models.MsgHistory{}
	json.Unmarshal(b, hist)

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

type YodelHandler struct {
	hub *ServerHub
}

func (y *YodelHandler) init() {
	y.hub.RegisterHandler(websocket_models.YodelCreate{}.Type(),
		y.HandleYodelCreate)
}

func (y *YodelHandler) HandleYodelCreate(b []byte, c *Client) {
	var yodel websocket_models.YodelCreate
	err := json.Unmarshal(b, &yodel)
	if err != nil {
		log.Printf("error in decoding yodelcreate json: %v", err)
		return
	}

	if yodel.Name == "" {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{
			Error:   "yodel_name_empty",
			Message: "Cannot create a server with no name!",
		}
		return
	}

	db_yodel := &database.Yodel{
		Name: yodel.Name,
	}
	
	err = y.hub.Database.InsertYodel(db_yodel)
    
	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
		return
	}

	c.OutgoingPayloadQueue <- websocket_models.Yodel{
		YodelID: db_yodel.YodelID.Hex(),
		Name: yodel.Name,
	}
}

func NewYodelHandler(hub *ServerHub) {
	y := YodelHandler{hub: hub}
	y.init()
}
