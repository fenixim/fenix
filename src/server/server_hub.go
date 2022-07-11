package server

import (
	"fenix/src/models"
	"fenix/src/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"

	"net/http"
)

type MainLoopEvent interface {
	GetEventType() string
}

type QuitMainLoop struct {}
func (e *QuitMainLoop) GetEventType() string {
	return "quit"
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
}

func (hub *ServerHub) Run(wg *utils.WaitGroupCounter) {
	err := wg.Add(1, "ServerHub_Run")
	if err != nil {
		log.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}
	defer wg.Done("ServerHub_Run")
	for {
		select {
		case client := <-hub.register:
			hub.clients[client.id] = client

		case client := <-hub.unregister:
			delete(hub.clients, client.id)
			client.ClientEventLoop <- ClientQuit{}

		case d := <-hub.broadcast:
			for _, client := range hub.clients {
				client.OutgoingMessageQueue <- d.ToJSON()
			}

		case e := <-hub.mainLoopEvent:
			log.Print("MainLoopYay")
			if e.GetEventType() == "quit" {
				for _, client := range hub.clients {
					client.ClientEventLoop <- ClientQuit{}
				}
				return
			}
		}
	}
}

func (hub *ServerHub) Upgrade(w http.ResponseWriter, r *http.Request) {
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

	client := &Client{hub: hub, conn: conn, nick: nick, id: uuid.NewString()}
	client.hub.register <- client
}

func Serve(addr string, wg *utils.WaitGroupCounter) *http.Server {
	srv := &http.Server{Addr: addr}
	hub := ServerHub{}
	NewHub()

	srv.RegisterOnShutdown(func() {
		hub.mainLoopEvent <- &QuitMainLoop{}
	})

	go hub.Run(wg)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Upgrade(w, r)
	})

	go func() {
		defer wg.Done("ServerHub_ListenAndServe")
		defer func() {
			hub.mainLoopEvent <- &QuitMainLoop{}
		}()
		
		err := wg.Add(1, "ServerHub_ListenAndServe")
		if err != nil {
			log.Fatalf("Error adding goroutine to waitgroup: %v", err)
		}
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv
}

func NewHub() *ServerHub {
	hub := ServerHub{}
	hub.clients = make(map[string]*Client)
	hub.register = make(chan *Client)
	hub.unregister = make(chan *Client)
	hub.broadcast = make(chan models.JSONModel)
	hub.mainLoopEvent = make(chan MainLoopEvent)
	return &hub
}
