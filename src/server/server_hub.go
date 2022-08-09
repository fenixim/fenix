package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fenix/src/database"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"fmt"
	"log"
	"sync"

	"net/http"

	"github.com/gorilla/websocket"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var version = "0.1"

// Main server class.  Should be initialized with NewHub()
type ServerHub struct {
	clients     *sync.Map
	broadcast_payload     chan websocket_models.JSONModel
	ctx         context.Context
	Shutdown    context.CancelFunc
	Handlers    map[string]func([]byte, *Client)
	Wg          *utils.WaitGroupCounter
	Database    database.Database
}

// Function to make and start an instance of ServerHub
func NewHub(wg *utils.WaitGroupCounter, database database.Database) *ServerHub {
	hub := ServerHub{
		clients: &sync.Map{},
		broadcast_payload: make(chan websocket_models.JSONModel),
		Handlers:  make(map[string]func([]byte, *Client)),
		Wg:        wg,
		Database:  database,
	}

	NewMessageHandler(&hub)
	NewIdentificationHandler(&hub)

	hub.ctx, hub.Shutdown = context.WithCancel(context.Background())

	go hub.run()

	return &hub
}

// Registers a message handler to be called when a type of message is recieved.
func (hub *ServerHub) RegisterHandler(messageType string, handler func([]byte, *Client)) {
	hub.Handlers[messageType] = handler
}

// Loop to broadcast payload to all clients.
// Will call callback "BroadcastPayload" when a request to broadcast a payload is made.
// Will call callback "BroadcastPayloadLoopDone" when this finishes.
func (hub *ServerHub) broadcast() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := hub.Wg.Add(1, "BroadcastPayloadLoop")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case d := <-hub.broadcast_payload:
				hub.clients.Range(func(key, value interface{}) bool {
					value.(*Client).OutgoingPayloadQueue <- d
					return true
				})


			case <-ctx.Done():
				hub.Wg.Done("BroadcastPayloadLoop")
				return
			}
		}
	}()

	return ctx, cancel
}

// Starts all goroutines for server to run.
// Will stop all goroutines when hub.Shutdown() is called.
func (hub *ServerHub) run() {
	_, broadcastCancel := hub.broadcast()

	err := hub.Wg.Add(1, "ServerHub_Run")
	if err != nil {
		panic(err)
	}

	<-hub.ctx.Done()
	hub.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		log.Printf("Closing client %v", client.User.UserID.Hex())
		client.Close("")
		return true
	})

	broadcastCancel()
	hub.Wg.Done("ServerHub_Run")
}

// Function to upgrade http connection to websocket
// Also makes new client.
func (hub *ServerHub) upgrade(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, http.Header{"Fenix-Version": []string{version}})
	if err != nil {
		log.Println(err)
		return
	}

	nick, _, _ := r.BasicAuth()

	client := &Client{hub: hub, conn: conn, User: database.User{Username: nick}}
	client.New()
	hub.clients.Store(client.User.UserID.Hex(), client)
}

// HTTP method to log in and upgrade a user's connection.
// Uses BasicAuth header
func (hub *ServerHub) Login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := &database.User{Username: username}

	err := hub.Database.GetUser(u)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	p := &database.User{Password: []byte(password), Salt: u.Salt}
	p.HashPassword()

	res := subtle.ConstantTimeCompare(p.Password, u.Password)
	if res != 1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	hub.upgrade(w, r)
}

func (hub *ServerHub) Register(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := &database.User{Username: username}
	err := hub.Database.GetUser(u)

	// If user exists, dont let the client re-register a user.
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	u.Salt = make([]byte, 16)

	rand.Read(u.Salt)
	u.Password = []byte(password)
	u.HashPassword()

	err = hub.Database.InsertUser(u)

	if err != nil {
		fmt.Errorf("Error inserting user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	hub.upgrade(w, r)
}

// Serves http server on addr.
func (hub *ServerHub) Serve(addr string) {
	srv := http.Server{
		Handler: hub.HTTPRequestHandler(),
		Addr:    addr,
	}

	defer hub.Wg.Done("ServerHub_ListenAndServe")

	err := hub.Wg.Add(1, "ServerHub_ListenAndServe")
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
func (hub *ServerHub) HTTPRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			hub.Login(w, r)
		} else if r.URL.Path == "/register" {
			hub.Register(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
