package server

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"fenix/src/models"
	"fenix/src/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/pbkdf2"
)

type MainLoopEvent interface {
	GetEventType() string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type hubChannels struct {
	broadcast     chan websocket_models.JSONModel
	mainLoopEvent chan MainLoopEvent
}

// Main server class.  Should be initialized with NewHub()
type ServerHub struct {
	MongoDatabase string
	clients       *sync.Map
	HubChannels   *hubChannels
	ctx           context.Context
	Shutdown      context.CancelFunc
	Handlers      map[string]func([]byte, *Client)
	callbacks     map[string]func([]interface{})
	Wg            *utils.WaitGroupCounter
	Database      *mongo.Client
}

// Function to make and start an instance of ServerHub
func NewHub(wg *utils.WaitGroupCounter) *ServerHub {
	hub := ServerHub{
		MongoDatabase: "development",
		clients:       &sync.Map{},
		HubChannels: &hubChannels{
			broadcast:     make(chan websocket_models.JSONModel),
			mainLoopEvent: make(chan MainLoopEvent),
		},
		Handlers:  make(map[string]func([]byte, *Client)),
		callbacks: make(map[string]func([]interface{})),
		Wg:        wg,
	}
	env, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(env["DB"]).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	hub.Database = c

	NewMessageHandler(&hub)
	NewIdentificationHandler(&hub)

	hub.ctx, hub.Shutdown = context.WithCancel(context.Background())

	go hub.Run(wg)

	return &hub
}

// Registers a message handler to be called when a type of message is recieved.
func (hub *ServerHub) RegisterHandler(messageType string, handler func([]byte, *Client)) {
	hub.Handlers[messageType] = handler
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
			case d := <-hub.HubChannels.broadcast:
				hub.clients.Range(func(key, value interface{}) bool {
					value.(*Client).OutgoingPayloadQueue <- d
					return true
				})

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
			case e := <-hub.HubChannels.mainLoopEvent:
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
	_, broadcastCancel := hub.Broadcast(wg)
	_, mainLoopEventsCancel := hub.MainLoopEvents(wg)

	err := wg.Add(1, "ServerHub_Run")
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

	nick, _, _ := r.BasicAuth()

	client := &Client{hub: hub, conn: conn, User: User{Username: nick}}
	client.New(wg)
	hub.clients.Store(client.User.UserID.Hex(), client)
}

// Serves http server on addr.
func Serve(addr string, wg *utils.WaitGroupCounter, hub *ServerHub) {
	srv := http.Server{
		Handler: HandleFunc(hub, wg),
		Addr:    addr,
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
		if r.URL.Path == "/ws" {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(401)
				return
			}

			u := User{Username: username}
			u.FindUser(hub)
			hash := pbkdf2.Key([]byte(password), u.Salt, 100000, 32, sha512.New512_256)

			res := subtle.ConstantTimeCompare(hash, u.Password)
			if res != 1 {
				w.WriteHeader(401)
				return
			}

			hub.Upgrade(w, r, wg)
		} else if r.URL.Path == "/register" {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(400)
				return
			}

			u := User{Username: username}

			u.UserID = primitive.NewObjectIDFromTimestamp(time.Now())

			err := u.FindUser(hub)
			if err == nil {
				w.WriteHeader(401)
				return
			}

			u.Salt = make([]byte, 16)

			rand.Read(u.Salt)
			u.Password = pbkdf2.Key([]byte(password), u.Salt, 100000, 32, sha512.New512_256)

			_, err = u.InsertUser(hub)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(500)
			}

			hub.Upgrade(w, r, wg)
		}

	}
}
