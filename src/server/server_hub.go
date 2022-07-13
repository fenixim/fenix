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

type ServerHub struct {
	clients       map[string]*Client
	register      chan *Client
	unregister    chan *Client
	broadcast     chan models.JSONModel
	mainLoopEvent chan MainLoopEvent
	ctx           context.Context
	Shutdown      context.CancelFunc
	handlers      map[string]func([]byte, *Client)
}

func (hub *ServerHub) RegisterHandler(messageType string, handler func([]byte, *Client)) {
	hub.handlers[messageType] = handler
}

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
			case <-ctx.Done():
				wg.Done("RegisterClientsLoop")
				return
			}
		}
	}()

	return ctx, cancel
}

func (hub *ServerHub) UnregisterClients(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "UnregisterClientsLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case client := <-hub.register:
				client.Close("")

			case <-ctx.Done():
				wg.Done("UnregisterClientsLoop")
				return
			}
		}
	}()

	return ctx, cancel
}

func (hub *ServerHub) Broadcast(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "BroadcastLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case d := <-hub.broadcast:
				for _, client := range hub.clients {
					client.OutgoingMessageQueue <- d
				}

			case <-ctx.Done():
				wg.Done("BroadcastLoop")
				return
			}
		}
	}()

	return ctx, cancel
}

func (hub *ServerHub) MainLoopEvents(wg *utils.WaitGroupCounter) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := wg.Add(1, "MainEventLoop")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case <-hub.mainLoopEvent:

			case <-ctx.Done():
				wg.Done("MainEventLoop")
				return
			}
		}
	}()

	return ctx, cancel
}

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
	client.hub.register <- client
}
func Init(wg *utils.WaitGroupCounter) *ServerHub {
	hub := NewHub()
	go hub.Run(wg)
	return hub
}

func Serve(addr *string, wg *utils.WaitGroupCounter) {
	srv := http.Server{
		Addr: *addr,
	}

	defer wg.Done("ServerHub_ListenAndServe")

	err := wg.Add(1, "ServerHub_ListenAndServe")
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func NewHub() *ServerHub {
	hub := ServerHub{
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan models.JSONModel),
		mainLoopEvent: make(chan MainLoopEvent),
	}

	NewMessageHandler(&hub)

	hub.ctx, hub.Shutdown = context.WithCancel(context.Background())

	return &hub
}

func HandleFunc(hub *ServerHub, wg *utils.WaitGroupCounter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hub.Upgrade(w, r, wg)
	}
}
