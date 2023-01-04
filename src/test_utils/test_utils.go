package test_utils

import (
	"context"
	"encoding/base64"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func AssertEqual(t *testing.T, got, expected interface{}) {
	t.Helper()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %v want %v", got, expected)
	}
}

func AssertNotEqual(t *testing.T, got, expected interface{}) {
	t.Helper()

	if reflect.DeepEqual(got, expected) {
		t.Errorf("got %q, didnt want %q", got, expected)
	}
}

type Credentials struct {
	Username string
	Password string
}

func RegisterClient(t *testing.T, srv *ServerFields, auth Credentials) *ClientFields {
	t.Helper()

	srv.Addr.Path = "/register"
	cli := Connect(auth.Username, auth.Password, srv.Addr)

	return cli
}

func LoginClient(t *testing.T, srv *ServerFields, auth Credentials) *ClientFields {
	t.Helper()

	srv.Addr.Path = "/login"
	cli := Connect(auth.Username, auth.Password, srv.Addr)

	return cli
}

type ServerFields struct {
	Database *database.InMemoryDatabase
	Wg       *utils.WaitGroupCounter
	Hub      *server.ServerHub
	Server   *httptest.Server
	Addr     url.URL
	Close    func()
}

type ClientFields struct {
	Conn  *websocket.Conn
	Res   *http.Response
	Close func()
}

func MsgHistory(t *testing.T, cli *ClientFields, from, to int64) {
	t.Helper()

	err := cli.Conn.WriteJSON(
		websocket_models.MsgHistory{From: from, To: to}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func MsgSend(t *testing.T, cli *ClientFields, content string) {
	t.Helper()

	err := cli.Conn.WriteJSON(
		websocket_models.MsgSend{Message: content}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func RecvMsgHistory(t *testing.T, cli *ClientFields) websocket_models.MsgHistory {
	t.Helper()

	var resProto websocket_models.MsgHistory
	err := cli.Conn.ReadJSON(&resProto)
	if err != nil {
		t.Fatal(err)
	}

	return resProto
}

func YodelCreate(t *testing.T, cli *ClientFields) {
	t.Helper()
	err := cli.Conn.WriteJSON(websocket_models.YodelCreate{}.SetType())
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func StartServer() *ServerFields {
	wg := utils.NewWaitGroupCounter()
	db := database.NewInMemoryDatabase()
	hub := server.NewHub(wg, db)

	srv := httptest.NewServer(hub.HTTPRequestHandler())
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	addr := url.URL{Scheme: "ws", Host: u.Host}
	return &ServerFields{
		Database: db,
		Wg:       wg,
		Hub:      hub,
		Server:   srv,
		Addr:     addr,
		Close: func() {
			hub.Shutdown()
			srv.Close()
		},
	}
}

func Connect(username, password string, u url.URL) *ClientFields {
	a := username + ":" + password
	auth := base64.StdEncoding.EncodeToString([]byte(a))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	conn, res, _ := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{"Authorization": []string{"Basic " + auth}})

	return &ClientFields{
		Conn: conn,
		Res:  res,
		Close: func() {
			if res.StatusCode == 101 {
				conn.Close()
			}
			cancel()
		},
	}
}

func StartServerAndConnect(username, password, endpoint string) (*ServerFields, *ClientFields, func()) {
	srv := StartServer()
	srv.Addr.Path = endpoint
	cli := Connect(username, password, srv.Addr)
	return srv, cli, func() {
		cli.Close()
		srv.Close()
	}
}
