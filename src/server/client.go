package server

import (
	"encoding/json"
	"fenix/src/database"
	"fenix/src/utils"
	"fenix/src/websocket_models"

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
	User                 database.User
	ClientEventLoop      chan ClientEvent
	OutgoingPayloadQueue chan websocket_models.JSONModel
}

// Can be called multiple times.  Should be deferred at end of functions
func (c *Client) Close(wg_id string) {
	if wg_id != "" {
		c.hub.Wg.Done(wg_id)
	}
	if !c.Closed {
		c.ClientEventLoop <- ClientQuit{}
	}

	c.Closed = true
	c.hub.Clients.Delete(c.User.UserID)

	c.conn.Close()
}

func (c *Client) New() {
	c.ClientEventLoop = make(chan ClientEvent)
	c.OutgoingPayloadQueue = make(chan websocket_models.JSONModel)

	c.User = database.User{Username: c.User.Username}
	c.hub.Database.GetUser(&c.User)

	c.conn.SetCloseHandler(c.OnClose)

	go c.listenOnEventLoop()
	go c.listenOnWebsocket()
}

func (c *Client) OnClose(code int, text string) error {
	c.ClientEventLoop <- ClientQuit{}
	c.Closed = true
	utils.InfoLogger.Printf("Client %v closed: Code %v, Reason %v", c.User.Username, code, text)
	return nil
}

func (c *Client) listenOnWebsocket() {
	err := c.hub.Wg.Add(1, "Client_ListenOnWebsocket__"+c.User.UserID.Hex())
	if err != nil {
		utils.ErrorLogger.Panicf("Error adding goroutine to waitgroup: %v", err)
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
			utils.InfoLogger.Printf("Error decoding message: %q; %q", err, b)
			c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "BadFormat", Message: "Error decoding: " + err.Error()}
			return
		}

		err = json.Unmarshal(b, &t)

		if err != nil {
			utils.InfoLogger.Printf("Error unmarshalling message: %q; %q", err, b)
			c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "BadFormat", Message: "Malformed JSON"}
			return
		}

		if handler, ok := c.hub.Handlers[t.Type]; ok {
			go handler(b, c)
		}
	}
}

func (c *Client) listenOnEventLoop() {
	err := c.hub.Wg.Add(1, "Client_ListenOnEventLoop__"+c.User.UserID.Hex())
	if err != nil {
		utils.ErrorLogger.Panicf("Error adding goroutine to waitgroup: %v", err)
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

			err := c.conn.WriteJSON(m.SetType())
			if err != nil {
				utils.WarningLogger.Printf("Error sending messsage of type %v to %v: %v", m.Type(), c.User.Username, err)
				c.Closed = true
				return
			}
		}
	}
}
