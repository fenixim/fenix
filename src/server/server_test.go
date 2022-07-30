package server

import (
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
			func(key, value interface{}) bool {
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
			func(key, value interface{}) bool {
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

	c := ConnectToServer(t, u.Host, "Gopher")

	c.Close()

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
