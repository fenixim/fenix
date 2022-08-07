package server

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageHandler struct {
	hub *ServerHub
}

func (m *MessageHandler) init() {
	m.hub.RegisterHandler("msg_send", m.HandleSendMessage)
	m.hub.RegisterHandler("msg_history", m.HandleMessageHistory)
}

func (m *MessageHandler) HandleSendMessage(b []byte, client *Client) {
	var msg SendMessage
	err := json.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("error in decoding message json: %v", err)
		return
	}

	recv_msg := BroadcastMessage{
		Time: time.Now().Unix(),
		Author: Author{
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
		client.OutgoingPayloadQueue <- GenericError{Error: "DatabaseError"}
	}
	recv_msg.MessageID = res.InsertedID.(primitive.ObjectID).Hex()

	m.hub.HubChannels.broadcast <- recv_msg
}

func (m *MessageHandler) HandleMessageHistory(b []byte, client *Client) {
	hist := &MessageHistory{}
	json.Unmarshal(b, hist)
	if hist.From == 0 || hist.To == 0 {
		client.OutgoingPayloadQueue <- GenericError{Error: "BadFormat", Message: "MessageHistory needs From and To fields"}
		return
	}

	messages, err := (&Message{}).GetMessagesBetween(m.hub, hist.From, hist.To)
	if err != nil {
		log.Printf("error in HandleMessageHistory, %v", err)
		return
	}
	hist.Messages = *messages
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
	c.OutgoingPayloadQueue <- WhoAmI{
		ID:   c.User.UserID.Hex(),
		Nick: c.User.Username,
	}
	i.hub.CallCallbackIfExists("WhoAmI", []interface{}{c})
}

func NewIdentificationHandler(hub *ServerHub) {
	i := IdentificationHandler{hub: hub}
	i.init()
}
