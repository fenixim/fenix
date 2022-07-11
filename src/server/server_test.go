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

func StartServer() (*utils.WaitGroupCounter, *http.Server) {
	wg := utils.WaitGroupCounter{}
	srv := Serve("localhost:8080", &wg)
	wg.Wait()
	return &wg, srv
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
	wg, srv := StartServer()

	start := wg.Counter
	c := ConnectToServer(t)
	c.Close()
	end := wg.Counter

	if start != end {
		t.Fail()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(3000))
	err := srv.Shutdown(ctx)
	cancel()
	if err != nil {
		t.Fatalf("Error shutting server down")
	}
}

func TestEnsureAllGoroutinesStopWhenServerExits(t *testing.T) {
	wg, srv := StartServer()
	ConnectToServer(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(3000))
	err := srv.Shutdown(ctx)
	cancel()
	if err != nil {
		t.Fatalf("Error shutting server down")
	}

	if wg.Counter != 0 {
		t.FailNow()
	}
}
