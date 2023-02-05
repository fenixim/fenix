package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fenix/src/database"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"sync"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var version = "0.1"

// Main server class.  Should be initialized with NewHub()
type ServerHub struct {
	Clients           *sync.Map
	Broadcast_payload chan websocket_models.JSONModel
	Ctx               context.Context
	Shutdown          context.CancelFunc
	Handlers          map[string]func([]byte, *Client)
	Wg                *utils.WaitGroupCounter
	Database          database.Database
	Tickets           *sync.Map
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
			case d := <-hub.Broadcast_payload:
				hub.Clients.Range(
					func(key, value interface{}) bool {
					go func() {value.(*Client).OutgoingPayloadQueue <- d}()
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
func (hub *ServerHub) Run() {
	_, broadcastCancel := hub.broadcast()

	err := hub.Wg.Add(1, "ServerHub_Run")
	if err != nil {
		panic(err)
	}

	<-hub.Ctx.Done()
	utils.InfoLogger.Printf("Server shutting down...")

	hub.Clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		client.Close("")
		return true
	})

	broadcastCancel()
	hub.Wg.Done("ServerHub_Run")
}

// Function to upgrade http connection to websocket
// Also makes new client.
func (hub *ServerHub) upgrade(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("t")
	userID := r.URL.Query().Get("id")

	if ticket == "" || userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vTicket, ok := hub.Tickets.LoadAndDelete(userID)

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	res := subtle.ConstantTimeCompare([]byte(vTicket.(string)), []byte(ticket))
	if res == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	conn, err := websocket.Upgrade(w, r, http.Header{}, 1024, 1024)
	if err != nil {
		utils.InfoLogger.Printf("Error upgrading connection to websocket: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.InfoLogger.Printf("Error parsing objectid: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u := database.User{UserID: id}
	hub.Database.GetUser(&u)
	client := &Client{hub: hub, conn: conn, User: database.User{Username: u.Username}}
	client.New()
	hub.Clients.Store(client.User.UserID.Hex(), client)
}

func (hub *ServerHub) createToken(u *database.User) []byte {
	ticket := make([]byte, 32)

	rand.Read(ticket)
	encodedTicket := base64.URLEncoding.EncodeToString(ticket)

	hub.Tickets.Store(u.UserID.Hex(), encodedTicket)
	go func() {
		time.Sleep(5 * time.Second)
		hub.Tickets.Delete(u.UserID)
	}()

	b, err := json.Marshal(map[string]interface{}{"userID": u.UserID.Hex(), "username": u.Username, "ticket": encodedTicket})
	if err != nil {
		utils.InfoLogger.Printf("Error marshalling JSON: %q", err)
		return nil
	}
	return b
}

// HTTP method to log in and upgrade a user's connection.
// Uses BasicAuth header
func (hub *ServerHub) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var body map[string]string
	err := decoder.Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	username, uok := body["username"]
	password, pok := body["password"]

	if !uok || !pok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := &database.User{Username: username}

	err = hub.Database.GetUser(u)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	p := &database.User{Password: []byte(password), Salt: u.Salt, Username: username}
	p.HashPassword()

	res := subtle.ConstantTimeCompare(p.Password, u.Password)
	if res != 1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	b := hub.createToken(u)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Write(b)
}

func (hub *ServerHub) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var body map[string]string
	err := decoder.Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	username, uok := body["username"]
	password, pok := body["password"]

	if !uok || !pok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := &database.User{Username: username}
	err = hub.Database.GetUser(u)

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
		utils.ErrorLogger.Printf("Error inserting user (%q:%q): %q", u.UserID.Hex(), u.Username, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b := hub.createToken(u)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Write(b)

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
		utils.ErrorLogger.Fatalf("Error adding goroutine to waitgroup: %v", err)
	}
	utils.InfoLogger.Printf("Listening on %v", addr)
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
		} else if r.URL.Path == "/upgrade" {
			hub.upgrade(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
