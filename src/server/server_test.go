package server_test

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

func TestWhoAmIOnWebsocket(t *testing.T) {
	t.Run("NormalWhoAmI", func(t *testing.T) {
		expectedUsername := "gopher1234"

		srv, cli, close := StartServerAndConnect(expectedUsername, "mytotallyrealpassword", "/register")

		defer close()

		cli.conn.WriteJSON(websocket_models.WhoAmI{}.SetType())
		var whoami websocket_models.WhoAmI

		cli.conn.ReadJSON(&whoami)

		gotUsername := whoami.Username
		if expectedUsername != gotUsername {
			t.Fail()
		}
		u := &database.User{Username: expectedUsername}
		srv.hub.Database.GetUser(u)
		expectedID := u.UserID.Hex()
		gotID := whoami.ID

		if expectedID != gotID {
			t.Fail()
		}
	})
}

func TestSendMessageOnWebsocket(t *testing.T) {
	compare := func(got, expected websocket_models.BroadcastMessage, t *testing.T) {
		if got.Author.ID != expected.Author.ID {
			t.Fail()
		}
		if got.Author.Username != expected.Author.Username {
			t.Fail()
		}
		if got.Message != expected.Message {
			t.Fail()
		}
	}
	t.Run("NormalSendMessage", func(t *testing.T) {
		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()
		user := database.User{Username: "gopher123"}
		srv.hub.Database.GetUser(&user)

		expectedMessage := "Hello there!\nGeneral Kenobi, you are a bold one!"
		expected := websocket_models.BroadcastMessage{
			Author: websocket_models.Author{
				ID:       user.UserID.Hex(),
				Username: user.Username,
			},
			Message: expectedMessage,
		}
		err := cli.conn.WriteJSON(websocket_models.SendMessage{
			Message: expectedMessage,
		}.SetType())

		if err != nil {
			panic(err)
		}
		var got websocket_models.BroadcastMessage
		cli.conn.ReadJSON(&got)

		compare(got, expected, t)
	})
	t.Run("EmptySendMessage", func(t *testing.T) {

		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()

		user := database.User{Username: "gopher123"}
		srv.hub.Database.GetUser(&user)

		expected := websocket_models.GenericError{
			Error: "message_empty",
		}

		cli.conn.WriteJSON(websocket_models.SendMessage{}.SetType())

		var got websocket_models.GenericError
		err := cli.conn.ReadJSON(&got)
		if err != nil {
			t.Fatal(err)
		}

		if got.Error != expected.Error {
			t.Fail()
		}
	})
}

func TestMessageHistoryOnWebsocket(t *testing.T) {
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

	t.Run("NormalMessageHistory", func(t *testing.T) {
		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()
		from := time.Now()
		populate(srv, 10)

		cli.conn.WriteJSON(websocket_models.MessageHistory{
			From: from.Unix(),
			To:   time.Now().Unix(),
		}.SetType())

		var got websocket_models.MessageHistory
		cli.conn.ReadJSON(&got)

		if len(got.Messages) != 10 {
			t.Fail()
		}
	})

	t.Run("EmptyMessageHistory", func(t *testing.T) {
		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()
		from := time.Now()
		populate(srv, 0)

		cli.conn.WriteJSON(websocket_models.MessageHistory{
			From: from.Unix(),
			To:   time.Now().Unix(),
		}.SetType())

		var got websocket_models.MessageHistory
		cli.conn.ReadJSON(&got)

		if len(got.Messages) != 0 {
			t.Fail()
		}
	})
	t.Run("FullMessageHistory", func(t *testing.T) {
		srv, cli, close := StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()
		from := time.Now()
		populate(srv, 100)

		cli.conn.WriteJSON(websocket_models.MessageHistory{
			From: from.Unix(),
			To:   time.Now().Unix(),
		}.SetType())

		var got websocket_models.MessageHistory
		cli.conn.ReadJSON(&got)

		if len(got.Messages) != 50 {
			t.Fail()
		}
	})

	t.Run("InvalidMessageInsert", func(t *testing.T) {
		_, cli, close := StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		cli.conn.WriteJSON(websocket_models.SendMessage{
			Message: "error",
		}.SetType())

		var got websocket_models.GenericError
		cli.conn.ReadJSON(&got)
		expected := websocket_models.GenericError{Error: "DatabaseError"}.SetType()

		if got != expected {
			t.Errorf("got %v expected %v", got, expected)
		}
	})
}
