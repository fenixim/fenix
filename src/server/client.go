package server

import (
	"fenix/src/models"
	"fenix/src/utils"
	"log"
	"time"

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
	hub    *ServerHub
	conn   *websocket.Conn
	nick   string
	id     string
	closed bool

	ClientEventLoop       chan ClientEvent
	OutgoingMessageQueue  chan models.JSONModel
	IncomingMessagesQueue chan models.JSONModel

	wg *utils.WaitGroupCounter
}

// Can be called multiple times.  Should be deferred at end of functions
func (c *Client) Close(wg_id string) {
	c.closed = true
	delete(c.hub.clients, c.id)

	c.conn.Close()

	if wg_id == "" {
		return
	}
	c.wg.Done(wg_id)
}

func (c *Client) New(wg *utils.WaitGroupCounter) {
	c.ClientEventLoop = make(chan ClientEvent)
	c.OutgoingMessageQueue = make(chan models.JSONModel)
	c.IncomingMessagesQueue = make(chan models.JSONModel)

	c.conn.SetCloseHandler(c.OnClose)
	err := c.conn.SetReadDeadline(time.Now().Add(time.Duration(5000)))

	if err != nil {
		log.Printf("error setting read deadline: %v", err)
	}

	c.wg = wg
	go c.listenOnEventLoop()
	go c.listenOnWebsocket()
}

func (c *Client) OnClose(code int, text string) error {
	c.ClientEventLoop <- ClientQuit{}
	c.closed = true
	log.Printf("Client %v closed: Code %v, Reason %v", c.nick, code, text)
	return nil
}

func (c *Client) listenOnWebsocket() {
	err := c.wg.Add(1, "Client_ListenOnWebsocket__"+c.id)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnWebsocket__" + c.id)

	for {
		var t models.JSONModel
		err := c.conn.ReadJSON(t)
		if err != nil {
			c.OutgoingMessageQueue <- models.BadFormat{Message: "Malformed JSON"}
			c.closed = true
			return
		}
		c.IncomingMessagesQueue <- t
	}
}

func (c *Client) listenOnEventLoop() {
	err := c.wg.Add(1, "Client_ListenOnEventLoop__"+c.id)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnEventLoop__" + c.id)

	select {
	case e := <-c.ClientEventLoop:
		if e.GetEventType() == "quit" {
			c.closed = true
			return
		}

	case m := <-c.OutgoingMessageQueue:
		if c.closed {
			return
		}
		err := c.conn.WriteJSON(m)
		if err != nil {
			log.Printf("Error sending messsage of type %v to %v: %v", m.Type(), c.nick, err)
			c.closed = true
			return
		}
	}
}
