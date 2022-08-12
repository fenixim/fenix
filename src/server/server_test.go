package server

import (
	"crypto/rand"
	"encoding/base64"
	"fenix/src/database"
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
	hub    *ServerHub
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
	wg := utils.WaitGroupCounter{}
	hub := NewHub(&wg, &database.StubDatabase{UsersById: &sync.Map{}, Messages: &sync.Map{}, UsersByUsername: &sync.Map{}})

	srv := httptest.NewServer(hub.HTTPRequestHandler())
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	addr := url.URL{Scheme: "ws", Host: u.Host}
	return &serverFields{
		wg:     &wg,
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
	conn, res, _ := websocket.DefaultDialer.Dial(u.String(), http.Header{"Authorization": []string{"Basic " + auth}})

	return &clientFields{
		conn: conn,
		res:  res,
		close: func() {
			if res.StatusCode == 101 {
				conn.Close()
			}
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

func TestEnsureCorrectStatusCodes(t *testing.T) {
	t.Run(
		"no auth header", func(t *testing.T) {
			srv := StartServer_()
			defer srv.close()

			srv.addr.Path = "/login"
			_, res, _ := websocket.DefaultDialer.Dial(srv.addr.String(), http.Header{})

			got := res.StatusCode
			expected := http.StatusBadRequest

			if got != expected {
				t.Fail()
			}
		},
	)
	t.Run(
		"user doesnt exist", func(t *testing.T) {
			_, cli, close := StartServerAndConnect("gopherboi", "gopher1234", "/login")
			defer close()
			got := cli.res.StatusCode
			expected := http.StatusForbidden

			if got != expected {
				t.Fail()
			}
		},
	)
	t.Run(
		"user does exist, wrong password", func(t *testing.T) {
			srv := StartServer_()
			defer srv.close()

			u := &database.User{
				Username: "gopher123",
				Salt:     make([]byte, 16),
				Password: []byte("gopher1234"),
			}

			rand.Read(u.Salt)
			u.HashPassword()
			srv.hub.Database.InsertUser(u)

			srv.addr.Path = "/login"
			cli := Connect_(u.Username, "notmypassword", srv.addr)
			got := cli.res.StatusCode
			expected := http.StatusForbidden

			if got != expected {
				t.Fail()
			}
		},
	)
	t.Run(
		"user does exist, right password", func(t *testing.T) {
			srv := StartServer_()
			defer srv.close()

			u := &database.User{
				Username: "gopher123",
				Salt:     make([]byte, 16),
				Password: []byte("gopher1234"),
			}
			rand.Read(u.Salt)
			u.HashPassword()
			srv.hub.Database.InsertUser(u)

			srv.addr.Path = "/login"
			cli := Connect_(u.Username, "gopher1234", srv.addr)
			got := cli.res.StatusCode
			expected := http.StatusSwitchingProtocols

			if got != expected {
				t.Fail()
			}
		},
	)
	t.Run(
		"user does not exist, register", func(t *testing.T) {
			srv, _, close := StartServerAndConnect("gopher1234", "go_is_great123", "/register")
			defer close()

			u := &database.User{
				Username: "gopher1234",
			}

			err := srv.hub.Database.GetUser(u)

			if err != nil {
				t.Fatalf("error getting user from db: %v", err)
			}
			got := u
			expected := &database.User{
				Username: "gopher1234",
			}

			if got.Username != expected.Username {
				t.Fail()
			}
		},
	)
	t.Run(
		"user does exist, register", func(t *testing.T) {
			srv := StartServer_()
			defer srv.close()

			u := &database.User{
				Username: "gopher123",
				Salt:     make([]byte, 16),
				Password: []byte("gopher1234"),
			}
			rand.Read(u.Salt)
			u.HashPassword()
			srv.hub.Database.InsertUser(u)

			srv.addr.Path = "/register"
			cli := Connect_(u.Username, "gopher1234", srv.addr)
			got := cli.res.StatusCode
			expected := http.StatusConflict

			if got != expected {
				t.Fail()
			}
		},
	)
}

func TestEnsureGoroutinesStop(t *testing.T) {
	t.Run("when client exits", func(t *testing.T) {
		srv := StartServer_()
		defer srv.close()
		srv.addr.Path = "/register"

		cli := Connect_("gopher123", "totallymypassword", srv.addr)
		cli.close()
		time.Sleep(10 * time.Millisecond)

		got := srv.wg.Counter
		if got != 2 {
			srv.wg.Names.Range(func(key, value interface{}) bool {
				t.Log(key)
				return true
			})
			t.Fail()
		}
	})
	t.Run("when server exits after client connected", func(t *testing.T) {
		srv := StartServer_()
		srv.addr.Path = "/register"
		cli := Connect_("gopher123", "totallymypassword", srv.addr)
		cli.close()
		srv.close()

		srv.wg.Wait()
		got := srv.wg.Counter
		expected := 0
		if got != expected {
			t.Fail()
		}
	})
	t.Run("when server exits", func(t *testing.T) {
		srv := StartServer_()
		srv.close()

		srv.wg.Wait()
		got := srv.wg.Counter
		expected := 0
		if got != expected {
			t.Fail()
		}
	})
	t.Run(
		"for incorrect path", func(t *testing.T) {
			_, cli, close := StartServerAndConnect("gopher123", "myawesomepassword", "/gophers")
			defer close()

			got := cli.res.StatusCode
			expected := http.StatusNotFound

			if got != expected {
				t.Fail()
			}
		},
	)
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

		cli.conn.WriteJSON(websocket_models.SendMessage{
			Message: expectedMessage,
		}.SetType())

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
