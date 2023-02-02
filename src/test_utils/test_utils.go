package test_utils

import (
	"context"
	"encoding/base64"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/server/runner"
	"fenix/src/utils"
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

type ServerFields struct {
	Database database.Database
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

func maybeGetMongoDB(mongoEnv ...map[string]string) database.Database {
	var db database.Database
	if len(mongoEnv) != 0 {
		addr, addrok := mongoEnv[0]["mongo_addr"]
		intTest, intTestOk := mongoEnv[0]["integration_testing"]
		if addrok && intTestOk {
			db = database.NewMongoDatabase(addr, intTest)
		}
		err := db.ClearDB()
		if err != nil {
			panic(err)
		}
	} else {
		db = database.NewInMemoryDatabase()
	}
	return db
}

func StartServer(mongoEnv ...map[string]string) *ServerFields {
	wg := utils.NewWaitGroupCounter()
	db := maybeGetMongoDB(mongoEnv...)

	hub := runner.NewHub(wg, db)

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

func StartServerAndConnect(username, password, endpoint string, mongoEnv ...map[string]string) (*ServerFields, *ClientFields, func()) {
	srv := StartServer(mongoEnv...)
	srv.Addr.Path = endpoint
	cli := Connect(username, password, srv.Addr)
	return srv, cli, func() {
		cli.Close()
		srv.Close()
	}
}

func PopulateDB(srv *ServerFields, count int) {
	user := database.User{Username: "gopher123"}
	srv.Hub.Database.GetUser(&user)

	for i := 0; i < count; i++ {
		srv.Hub.Database.InsertMessage(&database.Message{
			Content:   "Hello there!",
			Timestamp: time.Now().UnixNano(),
			Author:    user.UserID.Hex(),
		})
	}
}
