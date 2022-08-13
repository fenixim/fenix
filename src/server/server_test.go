package server_test

import (
	"context"
	"encoding/base64"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/test_utils"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type serverFields struct {
	wg     *utils.WaitGroupCounter
	hub    *server.ServerHub
	server *httptest.Server
	addr   url.URL
	close  func()
}

type clientFields struct {
	conn  *websocket.Conn
	res   *http.Response
	close func()
}

func askHistory(t *testing.T, cli *clientFields, from, to int64) {
	t.Helper()

	err := cli.conn.WriteJSON(
		websocket_models.MessageHistory{From: from, To: to}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func send(t *testing.T, cli *clientFields, content string) {
	t.Helper()

	err := cli.conn.WriteJSON(
		websocket_models.SendMessage{Message: content}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func recvHistory(t *testing.T, cli *clientFields) websocket_models.MessageHistory {
	t.Helper()

	var resProto websocket_models.MessageHistory
	err := cli.conn.ReadJSON(&resProto)
	if err != nil {
		t.Fatal(err)
	}

	return resProto
}

func StartServer_() *serverFields {
	wg := utils.NewWaitGroupCounter()
	hub := server.NewHub(wg, &database.StubDatabase{UsersById: &sync.Map{}, Messages: &sync.Map{}, UsersByUsername: &sync.Map{}})

	srv := httptest.NewServer(hub.HTTPRequestHandler())
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	addr := url.URL{Scheme: "ws", Host: u.Host}
	return &serverFields{
		wg:     wg,
		hub:    hub,
		server: srv,
		addr:   addr,
		close: func() {
			hub.Shutdown()
			srv.Close()
		},
	}
}

func Connect_(username, password string, u url.URL) *clientFields {
	a := username + ":" + password
	auth := base64.StdEncoding.EncodeToString([]byte(a))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	conn, res, _ := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{"Authorization": []string{"Basic " + auth}})

	return &clientFields{
		conn: conn,
		res:  res,
		close: func() {
			if res.StatusCode == 101 {
				conn.Close()
			}
			cancel()
		},
	}
}

func StartServerAndConnect(username, password, endpoint string) (*serverFields, *clientFields, func()) {
	srv := StartServer_()
	srv.addr.Path = endpoint
	cli := Connect_(username, password, srv.addr)
	return srv, cli, func() {
		cli.close()
		srv.close()
	}
}

func TestProtocols(t *testing.T) {
	t.Run("whoami", func(t *testing.T) {
		_, cli, closeConn := StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		cli.conn.WriteJSON(websocket_models.WhoAmI{}.SetType())

		var resProto websocket_models.WhoAmI
		cli.conn.ReadJSON(&resProto)

		expected := "gopher123"
		got := resProto.Username
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("broadcast message", func(t *testing.T) {
		_, cli, closeConn := StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		send(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.BroadcastMessage
		cli.conn.ReadJSON(&resProto)

		got := resProto.Message
		expected := "General Kenobi, you are a bold one!"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("broadcast username", func(t *testing.T) {
		_, cli, close := StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		send(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.BroadcastMessage
		cli.conn.ReadJSON(&resProto)

		got := resProto.Author.Username
		expected := "gopher123"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("send empty message", func(t *testing.T) {

		_, cli, closeConn := StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		send(t, cli, "")

		var resProto websocket_models.GenericError
		err := cli.conn.ReadJSON(&resProto)
		if err != nil {
			t.Fatal(err)
		}

		expected := "message_empty"
		got := resProto.Error
		test_utils.AssertEqual(t, got, expected)
	})

	populate := func(srv *serverFields, count int) {
		user := database.User{Username: "gopher123"}
		srv.hub.Database.GetUser(&user)

		for i := 0; i < count; i++ {
			srv.hub.Database.InsertMessage(&database.Message{
				Content:   "Hello there!",
				Timestamp: time.Now().Unix(),
				Author:    user.UserID.Hex(),
			})
		}
	}

	t.Run("message history length", func(t *testing.T) {
		srv, cli, closeConn := StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		populate(srv, 1)
		askHistory(t, cli, 0, time.Now().Unix())

		got := len(recvHistory(t, cli).Messages)
		expected := 1
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("empty message history length", func(t *testing.T) {
		_, cli, closeConn := StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		askHistory(t, cli, 0, time.Now().Unix())

		got := len(recvHistory(t, cli).Messages)
		expected := 0
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("message history limit length", func(t *testing.T) {
		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()

		populate(srv, 51)
		askHistory(t, cli, 0, time.Now().Unix())

		got := len(recvHistory(t, cli).Messages)
		expected := 50
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("database error", func(t *testing.T) {
		_, cli, close := StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		send(t, cli, "error")

		var resProto websocket_models.GenericError
		cli.conn.ReadJSON(&resProto)

		got := resProto.Error
		expected := "DatabaseError"
		test_utils.AssertEqual(t, got, expected)
	})
}
