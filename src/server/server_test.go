package server

import (
	"context"
	"encoding/json"
	"fenix/src/models"
	"fenix/src/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func StartServer() (*utils.WaitGroupCounter, *httptest.Server, *ServerHub) {
	wg := utils.NewWaitGroupCounter()
	hub := Init(wg)

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

func TestEnsureAllGoroutinesStopWhenClientExits(t *testing.T) {
	wg, srv, hub := StartServer()

	// The server apparently doesnt start serving until a connection is made.
	// This is just to initialize the server loop and has no effect on the count,
	// other than making it include the needed server loop
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	c := ConnectToServer(t, u.Host, "Gopher")
	c.Close()
	time.Sleep(10 * time.Millisecond)

	start := wg.Counter

	c = ConnectToServer(t, u.Host, "Gopher2")
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
	wg, srv, hub := StartServer()

	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	ConnectToServer(t, u.Host, "Gopher")

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
	_, srv, hub := StartServer()

	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	 ConnectToServer(t, u.Host, "Gopher")

	// c.Close()

	time.Sleep(10 * time.Millisecond)

	keys := make([]string, 0, len(hub.clients))
	for k := range hub.clients {
		keys = append(keys, k)
	}

	if len(keys) != 0 {
		t.Logf("%v", keys)
		t.Logf("Did not delete client!")
		t.Fail()
	}

	hub.Shutdown()
	srv.Close()
}

func TestEnsureMessagesWork(t *testing.T) {
	_, srv, hub := StartServer()

	defer srv.Close()
	defer hub.Shutdown()
	
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	c := ConnectToServer(t, u.Host, "Gopher")

	defer c.Close()
	
	log.Printf("clients has %v", len(hub.clients))

	clients := make([]*Client, 0, len(hub.clients))
	for _, cli := range hub.clients {
		clients = append(clients, cli)
	}

	sendMessage := models.SendMessage{T: "send_message", Message: "Hello there"}
	timeSent := time.Now()
	err = c.WriteJSON(sendMessage)
	if err != nil {
		t.Fatalf("Error writing JSON: %v", err)
	}

	messageTimeout, cancel  := context.WithTimeout(context.Background(), time.Second * 5)

	select {
		case <- clients[0].IncomingMessagesQueue:
			
		case <- messageTimeout.Done():
			t.Log("Server did not recieve message in 5 seconds.")
			t.Fail()
	}
	cancel()

	
	_, b, err := c.ReadMessage()

	if err != nil {
		t.Fatalf("Error reading message from server: %v", err)
	}

	var j models.JSONModel
	json.Unmarshal(b, j)
	
	if j.Type() != "recv_msg" {
		t.Fatalf("%v", b)
	}

	var recvMsg models.RecvMessage
	json.Unmarshal(b, recvMsg)

	if recvMsg.Author != clients[0].Nick {
		t.Logf("%v's message shown as %v", clients[0].Nick, recvMsg.Author)
		t.Fail()
	}
	if recvMsg.Message != sendMessage.Message {
		t.Logf("Message shown as %v", recvMsg.Message)
	}
	if (recvMsg.Time - timeSent.Unix()) > int64(time.Second) {
		t.Logf("Time diff over 1 sec: %v", (recvMsg.Time - timeSent.Unix()))
	}
}