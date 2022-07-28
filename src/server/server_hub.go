package server

import (
	"context"
	"fenix/src/models"
	"fenix/src/utils"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"net/http"
)

type MainLoopEvent interface {
	GetEventType() string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Main server class.  Should be initialized with NewHub() 
type ServerHub struct {
	clients       map[string]*Client
	register      chan *Client
	unregister    chan *Client
	broadcast     chan models.JSONModel
	mainLoopEvent chan MainLoopEvent
	ctx           context.Context
	Shutdown      context.CancelFunc
	handlers      map[string]func([]byte, *Client)
	callbacks     map[string]func([]interface{})
}

// Function to make and start an instance of ServerHub
func NewHub(wg *utils.WaitGroupCounter) *ServerHub {
	hub := ServerHub{
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan models.JSONModel),
		mainLoopEvent: make(chan MainLoopEvent),
		handlers:      make(map[string]func([]byte, *Client)),
		callbacks:     make(map[string]func([]interface{})),
	}
	
	NewMessageHandler(&hub)
	hub.RegisterHandler("whoami", hub.WhoAmI)

	hub.ctx, hub.Shutdown = context.WithCancel(context.Background())

	go hub.Run(wg)

	return &hub
}

// Registers a message handler to be called when a type of message is recieved.
func (hub *ServerHub) RegisterHandler(messageType string, handler func([]byte, *Client)) {
	hub.handlers[messageType] = handler
}

// Registers a callback to be called when event happens.  If a function calls a callback, it should be shown in a docstring.
func (hub *ServerHub) RegisterCallback(event string, f func([]interface{})) {
	hub.callbacks[event] = f
}

// Helper function to reduce boilerplate code for calling callbacks.
func (hub *ServerHub) CallCallbackIfExists(name string, args []interface{}) {
	if callback, ok := hub.callbacks[name]; ok {
		callback(args)
	}
}

// Loop to register clients.
// Will call callback "RegisterClient" when a request to register a client is made
// Will call callback "RegisterClientLoopDone" when this finishes.
func (hub *ServerHub) RegisterClients(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "RegisterClientsLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.clients[client.ID] = client
				hub.CallCallbackIfExists("RegisterClient", []interface{}{client})

			case <-ctx.Done():
				wg.Done("RegisterClientsLoop")
				hub.CallCallbackIfExists("RegisterClientLoopDone", []interface{}{})
				return
			}
		}
	}()

	return ctx, cancel
}

// Loop to unregister clients.
// Will call callback "UnregisterClient" when a request to unregister a client is made
// Will call callback "UnregisterClientLoopDone" when this finishes.
func (hub *ServerHub) UnregisterClients(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "UnregisterClientsLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case client := <-hub.unregister:
				client.Close("")
				hub.CallCallbackIfExists("UnregisterClient", []interface{}{client})

			case <-ctx.Done():
				wg.Done("UnregisterClientsLoop")
				hub.CallCallbackIfExists("UnregisterClientLoopDone", []interface{}{})

				return
			}
		}
	}()

	return ctx, cancel
}

// Loop to broadcast payload to all clients.
// Will call callback "BroadcastPayload" when a request to broadcast a payload is made.
// Will call callback "BroadcastPayloadLoopDone" when this finishes.
func (hub *ServerHub) Broadcast(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "BroadcastPayloadLoop")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case d := <-hub.broadcast:
				for _, client := range hub.clients {
					client.OutgoingPayloadQueue <- d
				}
				hub.CallCallbackIfExists("BroadcastPayload", []interface{}{d})

			case <-ctx.Done():
				wg.Done("BroadcastPayloadLoop")
				hub.CallCallbackIfExists("BroadcastPayloadLoopDone", []interface{}{})
				return
			}
		}
	}()

	return ctx, cancel
}

// Loop to recieve main loop commands.  Currently is unused, but possibly will be used in the future.
// Will call callback "MainLoopEvent" when a main loop event is dispatched.
// Will call callback "MainEventLoopDone" when this finishes.
func (hub *ServerHub) MainLoopEvents(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "MainEventLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case e := <-hub.mainLoopEvent:
				hub.CallCallbackIfExists("MainLoopEvent", []interface{}{e})
				
			case <-ctx.Done():
				wg.Done("MainEventLoop")
				hub.CallCallbackIfExists("MainEventLoopDone", []interface{}{})
				return
			}
		}
	}()

	return ctx, cancel
}

// Starts all goroutines for server to run.
// Will stop all goroutines when hub.Shutdown() is called.
func (hub *ServerHub) Run(wg *utils.WaitGroupCounter) {
	_, registerCancel := hub.RegisterClients(wg)
	_, unregisterCancel := hub.UnregisterClients(wg)
	_, broadcastCancel := hub.Broadcast(wg)
	_, mainLoopEventsCancel := hub.MainLoopEvents(wg)

	err := wg.Add(1, "ServerHub_Run")
	if err != nil {
		panic(err)
	}

	<-hub.ctx.Done()
	for _, client := range hub.clients {
		log.Printf("Closing client %v", client.ID)
		client.Close("")
	}
	registerCancel()
	unregisterCancel()
	broadcastCancel()
	mainLoopEventsCancel()
	wg.Done("ServerHub_Run")
}

// Function to upgrade http connection to websocket
// Also makes new client.
func (hub *ServerHub) Upgrade(w http.ResponseWriter, r *http.Request, wg *utils.WaitGroupCounter) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var nick string
	if n, ok := r.Header["Nick"]; ok {
		nick = n[0]
	} else {
		nick = uuid.NewString()
	}

	client := &Client{hub: hub, conn: conn, Nick: nick, ID: uuid.NewString()}
	client.New(wg)
	hub.register <- client
}

// Handler for whoami requests
func (hub *ServerHub) WhoAmI(_ []byte, c *Client) {
	c.OutgoingPayloadQueue <- models.WhoAmI{
		T:    "whoami",
		ID:   c.ID,
		Nick: c.Nick,
	}
	hub.CallCallbackIfExists("WhoAmI", []interface{}{c})
}

// Serves http server on addr.
func Serve(addr string, wg *utils.WaitGroupCounter, hub *ServerHub) {
	srv := http.Server{
		Handler: HandleFunc(hub, wg),
		Addr: addr,
	}

	defer wg.Done("ServerHub_ListenAndServe")

	err := wg.Add(1, "ServerHub_ListenAndServe")
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}
	log.Printf("Listening on %v", addr)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// Handler func for incoming requests.
func HandleFunc(hub *ServerHub, wg *utils.WaitGroupCounter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hub.Upgrade(w, r, wg)
	}
}
