package server

import (
	"encoding/json"
	websocket_models "fenix/src/models"
	"fenix/src/utils"
	"fmt"
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

// Representation of the client for the server.  Spawns its own goroutines for message processing.
type Client struct {
	hub                  *ServerHub
	conn                 *websocket.Conn
	Closed               bool
	User                 User
	ClientEventLoop      chan ClientEvent
	OutgoingPayloadQueue chan websocket_models.JSONModel

	wg *utils.WaitGroupCounter
}

// Can be called multiple times.  Should be deferred at end of functions
func (c *Client) Close(wg_id string) {
	if wg_id != "" {
		c.wg.Done(wg_id)
	}
	c.Closed = true
	c.hub.clients.Delete(c.User.UserID)

	c.conn.Close()


}

func (c *Client) New(wg *utils.WaitGroupCounter) {
	c.ClientEventLoop = make(chan ClientEvent)
	c.OutgoingPayloadQueue = make(chan websocket_models.JSONModel)

	c.User = User{Username: c.User.Username}
	c.User.FindUser(c.hub)

	c.conn.SetCloseHandler(c.OnClose)

	c.wg = wg
	go c.listenOnEventLoop()
	go c.listenOnWebsocket()
}

func (c *Client) OnClose(code int, text string) error {
	c.ClientEventLoop <- ClientQuit{}
	c.Closed = true
	log.Printf("Client %v closed: Code %v, Reason %v", c.User.Username, code, text)
	return nil
}

func (c *Client) listenOnWebsocket() {
	err := c.wg.Add(1, "Client_ListenOnWebsocket__" + c.User.UserID.Hex())
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnWebsocket__" + c.User.UserID.Hex())

	for {
		var t struct {
			Type string `json:"type"`
		}
		_, b, err := c.conn.ReadMessage()
		
		if websocket.IsUnexpectedCloseError(err) {
			return
		}

		if err != nil {
			fmt.Println(err)
			c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "BadFormat", Message: "Error decoding: " + err.Error()}
			return
		}

		err = json.Unmarshal(b, &t)

		if err != nil {
			c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "BadFormat", Message: "Malformed JSON"}
			c.Closed = true
			return
		}

		if handler, ok := c.hub.Handlers[t.Type]; ok {
			go handler(b, c)
		}
	}
}

func (c *Client) listenOnEventLoop() {
	err := c.wg.Add(1, "Client_ListenOnEventLoop__"+c.User.UserID.Hex())
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	defer c.Close("Client_ListenOnEventLoop__" + c.User.UserID.Hex())
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
				log.Printf("Error sending messsage of type %v to %v: %v", m.Type(), c.User.Username, err)
				c.Closed = true
				return
			}
		}
	}
}
