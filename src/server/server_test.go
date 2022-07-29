package server

import (
	"encoding/json"
	"fenix/src/models"
	"fenix/src/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func StartServer() (*utils.WaitGroupCounter, *httptest.Server, *ServerHub) {
	wg := utils.NewWaitGroupCounter()
	hub := NewHub(wg)

	srv := httptest.NewServer(HandleFunc(hub, wg))
	return wg, srv, hub
}

func ConnectToServer(t *testing.T, host string, nick string) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	nick_header := make([]string, 1)
	nick_header[0] = nick
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Nick": nick_header})
	if err != nil {
		t.Fatalf("Error connecting to localhost:8080: %v", err)
	}
	return c
}

func Prepare() (*utils.WaitGroupCounter, *httptest.Server, *ServerHub, string){
	wg, srv, hub := StartServer()

	// The server apparently doesnt start serving until a connection is made.
	// This is just to initialize the server loop and has no effect on the count,
	// other than making it include the needed server loop
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}
	return wg, srv, hub, u.Host
}

func SendOnWebsocket(conn *websocket.Conn, payload models.JSONModel, t *testing.T) {
	err := conn.WriteJSON(payload)
	if err != nil {
		t.Fatalf("Error sending payload: %v", err)
	}
}

func RecvOnWebsocket(conn *websocket.Conn, t *testing.T) ([]byte) {
	_, b, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Error recieving message: %v", err)
	}
	return b
}

func TestEnsureAllGoroutinesStopWhenClientExits(t *testing.T) {
	wg, srv, hub, addr := Prepare()

	c := ConnectToServer(t, addr, "Gopher")
	c.Close()
	time.Sleep(10 * time.Millisecond)

	start := wg.Counter

	c = ConnectToServer(t, addr, "Gopher2")
	c.Close()
	time.Sleep(10 * time.Millisecond)
	end := wg.Counter

	if start != end {
		t.Log("failed, goroutines still running:")
		t.Logf("Start: %v, End: %v", start, end)
		wg.Names.Range(
			func(key, value any) bool {
				t.Log(key)
				return true
			},
		)
		t.FailNow()
	}

	hub.Shutdown()
	srv.Close()
}

func TestEnsureAllGoroutinesStopWhenServerExits(t *testing.T) {
	wg, srv, hub, addr := Prepare()
	ConnectToServer(t, addr, "Gopher")

	hub.Shutdown()
	srv.Close()

	wg.Wait()

	if wg.Counter != 0 {
		t.Log("failed, goroutines still running:")
		wg.Names.Range(
			func(key, value any) bool {
				t.Log(key)
				return true
			},
		)
		t.FailNow()
	}
}

func TestEnsureClientIsDeletedWhenDisconnected(t *testing.T) {
	_, srv, hub, addr := Prepare()

	defer hub.Shutdown()
	defer srv.Close()

	c := ConnectToServer(t, addr, "Gopher")
	c.Close()

	time.Sleep(10 * time.Millisecond)

	keys := make([]any, 0)

	hub.clients.Range(func(key, value any) bool {
		keys = append(keys, key)
		return true
	})

	if len(keys) != 0 {
		t.Logf("%v", keys)
		t.Logf("Did not delete client!")
		t.Fail()
	}
}

func TestSendPayloadOnWebsocket(t *testing.T) {
	_, srv, hub, addr := Prepare()

	defer hub.Shutdown()
	defer srv.Close()

	conn := ConnectToServer(t, addr, "Gopher")
	defer conn.Close()

	SendOnWebsocket(conn, models.SendMessage{T: "send_message", Message: "Hello world!"}, t)
}

func TestRecievePayloadOnWebsocket(t *testing.T) {
	_, srv, hub, addr := Prepare()

	defer hub.Shutdown()
	defer srv.Close()

	conn := ConnectToServer(t, addr, "Gopher")
	defer conn.Close()

	time.Sleep(10 * time.Millisecond)

	clients := make([]*Client, 0)
	hub.clients.Range(func(key, value any) bool {
		clients = append(clients, value.(*Client))
		return true
	})

	client := clients[0]

	hub.handlers["whoami"](make([]byte, 0), client)

	w := &models.WhoAmI{}
	b := RecvOnWebsocket(conn, t)
	err := json.Unmarshal(b, w)
	if err != nil {
		t.FailNow()
	}
}


func TestWhoAmI(t *testing.T) {
	_, srv, hub, addr := Prepare()

	defer srv.Close()
	defer hub.Shutdown()

	conn := ConnectToServer(t, addr, "Gopher")

	defer conn.Close()
	
	SendOnWebsocket(conn, models.WhoAmI{T: "whoami"}, t)

	var w models.WhoAmI
	b := RecvOnWebsocket(conn, t)
	err := json.Unmarshal(b, &w)
	if err != nil {
		t.FailNow()
	}
	if w.Nick != "Gopher" || w.ID == "" {
		t.Fatalf("nick: %v, id: %v", w.Nick, w.ID)
	}
}
