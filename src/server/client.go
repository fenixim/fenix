package server

import (
	"fenix/src/models"
	"fenix/src/utils"
	"log"

	"github.com/gorilla/websocket"
)

type ClientEvent interface {
	GetEventType() string
}

type ClientQuit struct{}

func (c ClientQuit) GetEventType() string {
	return "quit"
}

type Client struct {
	hub             *ServerHub
	conn            *websocket.Conn
	nick            string
	id              string
	ClientEventLoop chan ClientEvent

	OutgoingMessageQueue  chan *models.MessageType
	IncomingMessagesQueue chan *models.MessageType

	wg *utils.WaitGroupCounter
}

func (c *Client) New(wg *utils.WaitGroupCounter) {
	c.ClientEventLoop = make(chan ClientEvent)
	c.conn.SetCloseHandler(c.Close)

	go c.listenOnEventLoop()
	go c.listenOnWebsocket()
}

func (c *Client) Close(code int, text string) error {
	c.ClientEventLoop <- ClientQuit{}
	log.Printf("Client %v closed: Code %v, Reason %v", c.nick, code, text)
	return nil
}

func (c *Client) Send(d models.JSONModel) {
	err := c.conn.WriteJSON(d.ToJSON())
	if err != nil {
		log.Printf("error sending message: %v", err)
	}
}

func (c *Client) listenOnWebsocket() {
	err := c.wg.Add(1, "Client_ListenOnWebsocket__"+c.id)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}
	defer c.wg.Done("Client_ListenOnWebsocket__" + c.id)
	defer c.conn.Close()

	for {
		var t models.MessageType
		err := c.conn.ReadJSON(t)
		if err != nil {
			c.OutgoingMessageQueue <- models.BadFormat{Message: "Malformed JSON"}.ToJSON()
			return
		}
		c.IncomingMessagesQueue <- &t
	}
}

func (c *Client) listenOnEventLoop() {
	err := c.wg.Add(1, "Client_ListenOnEventLoop__"+c.id)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}
	defer c.wg.Done("Client_ListenOnEventLoop__" + c.id)
	defer c.conn.Close()

	select {
	case e := <-c.ClientEventLoop:
		if e.GetEventType() == "quit" {
			return
		}

	case m := <-c.OutgoingMessageQueue:
		err := c.conn.WriteJSON(m)
		if err != nil {
			log.Printf("Error sending messsage of type %v to %v: %v", m.MessageType, c.nick, err)
		}
	}
}
