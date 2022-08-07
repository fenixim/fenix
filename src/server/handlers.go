package server

import (
	"context"
	"encoding/json"
	websocket_models "fenix/src/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageHandler struct {
	hub *ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("msg_send", m.HandleSendMessage)
}

func (m *MessageHandler) HandleSendMessage(b []byte, client *Client) {
	var msg websocket_models.SendMessage
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	recv_msg := websocket_models.BroadcastMessage{
		T:    "msg_broadcast",
		Time: time.Now().Unix(),
		Author: websocket_models.Author{
			ID:   client.User.UserID.Hex(),
			Nick: client.User.Username,
		},
		Message: msg.Message,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	db_msg := Message{
		MessageID: primitive.NewObjectIDFromTimestamp(time.Unix(recv_msg.Time,0)),
		Content: recv_msg.Message,
		Timestamp: recv_msg.Time,
		Author: recv_msg.Author.ID,
	}

	res, err := m.hub.Database.Database(m.hub.MongoDatabase).Collection("messages").InsertOne(ctx, db_msg)

	if err != nil {
		client.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
	}
	recv_msg.MessageID = res.InsertedID.(primitive.ObjectID).Hex()

	m.hub.HubChannels.broadcast <- recv_msg
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
		T:    "whoami",
		ID:   c.User.UserID.Hex(),
		Nick: c.User.Username,
	}
	i.hub.CallCallbackIfExists("WhoAmI", []interface{}{c})
}

func NewIdentificationHandler(hub *ServerHub) {
	i := IdentificationHandler{hub: hub}
	i.init()
}
