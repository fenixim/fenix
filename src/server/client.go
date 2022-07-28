package server

import (
	"encoding/json"
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
	hub    *ServerHub
	conn   *websocket.Conn
	Nick   string
	ID     string
	Closed bool

	ClientEventLoop      chan ClientEvent
	OutgoingPayloadQueue chan models.JSONModel
	// May take out next version, not currently used.  Would not impact number of goroutines / client
	IncomingMessagesQueue chan models.JSONModel

	wg *utils.WaitGroupCounter
}

// Can be called multiple times.  Should be deferred at end of functions
func (c *Client) Close(wg_id string) {
	c.Closed = true
	delete(c.hub.clients, c.ID)

	c.conn.Close()

	if wg_id == "" {
		return
	}
	c.wg.Done(wg_id)
}

func (c *Client) New(wg *utils.WaitGroupCounter) {
	c.ClientEventLoop = make(chan ClientEvent)
	c.OutgoingPayloadQueue = make(chan models.JSONModel)
	c.IncomingMessagesQueue = make(chan models.JSONModel)

	c.conn.SetCloseHandler(c.OnClose)

	c.wg = wg
	go c.listenOnEventLoop()
	go c.listenOnWebsocket()
}

func (c *Client) OnClose(code int, text string) error {
	c.ClientEventLoop <- ClientQuit{}
	c.Closed = true
	log.Printf("Client %v closed: Code %v, Reason %v", c.Nick, code, text)
	return nil
}

func (c *Client) listenOnWebsocket() {
	err := c.wg.Add(1, "Client_ListenOnWebsocket__"+c.ID)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnWebsocket__" + c.ID)

	for {
		var t map[string]interface{}
		_, p, err := c.conn.ReadMessage()

		// if len(p) == 0 {
			// continue
		// }

		if err != nil {
			c.OutgoingPayloadQueue <- models.BadFormat{Message: "Error decoding: " + err.Error()}
			return
		}
		err = json.Unmarshal(p, &t)

		if err != nil {
			c.OutgoingPayloadQueue <- models.BadFormat{Message: "Malformed JSON"}
			c.Closed = true
			return
		}

		if handler, ok := c.hub.handlers[t["type"].(string)]; ok {
			go handler(p, c)
		}
	}
}

func (c *Client) listenOnEventLoop() {
	err := c.wg.Add(1, "Client_ListenOnEventLoop__"+c.ID)
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnEventLoop__" + c.ID)
	for {
		select {
		case e := <-c.ClientEventLoop:
			if e.GetEventType() == "quit" {
				c.Closed = true
				return
			}

		case m := <-c.OutgoingPayloadQueue:
			if c.Closed {
				return
			}
			err := c.conn.WriteJSON(m)
			if err != nil {
				log.Printf("Error sending messsage of type %v to %v: %v", m.Type(), c.Nick, err)
				c.Closed = true
				return
			}
		}
	}
}
