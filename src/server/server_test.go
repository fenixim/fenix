package server

import (
	"context"
	"fenix/src/utils"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func StartServer() (*utils.WaitGroupCounter, *http.Server, *ServerHub) {
	wg := utils.NewWaitGroupCounter()
	srv, hub := Serve("localhost:8080", wg)
	wg.Wait()
	return wg, srv, hub
}

func ConnectToServer(t *testing.T) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
	c := ConnectToServer(t)
	c.Close()
	time.Sleep(10 * time.Millisecond)

	start := wg.Counter
	c = ConnectToServer(t)
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

	hub.mainLoopEvent <- &QuitMainLoop{}
	err := srv.Shutdown(context.TODO())
	if err != nil {
		t.Fatalf("Error shutting server down: %v", err)
	}
}

func TestEnsureAllGoroutinesStopWhenServerExits(t *testing.T) {
	wg, srv, hub := StartServer()
	ConnectToServer(t)

	hub.mainLoopEvent <- &QuitMainLoop{}
	err := srv.Shutdown(context.TODO())

	if err != nil {
		t.Fatalf("Error shutting server down: %v", err)
	}

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
