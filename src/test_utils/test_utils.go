package test_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/server/runner"
	"fenix/src/utils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

func intTestDB() database.Database {
	var db database.Database
	mongoAddr := os.Getenv("mongo_addr")
	intTest := os.Getenv("integration_testing")

	if mongoAddr == "" || intTest == "" {
		utils.ErrorLogger.Panicf("Couldn't get database env -  mongoAddr: %q   intTest: %q", mongoAddr, intTest)
	} else {
		db = database.NewMongoDatabase(mongoAddr, intTest)
		err := db.ClearDB()
		if err != nil {
			panic(err)
		}
	}
	return db
}

func StartServer(isIntTest ...bool) *ServerFields {
	utils.InitLogger(3)
	wg := utils.NewWaitGroupCounter()
	var db database.Database

	if len(isIntTest) == 1 && isIntTest[0] {
		db = intTestDB()
	} else {
		db = database.NewInMemoryDatabase()
	}

	hub := runner.NewHub(wg, db)

	srv := httptest.NewServer(hub.HTTPRequestHandler())
	u, err := url.ParseRequestURI(srv.URL)
	if err != nil {
		panic(err)
	}

	addr := url.URL{Scheme: "http", Host: u.Host}
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
	b, err := json.Marshal(map[string]string{"username": username, "password": password})
	if err != nil {
		panic(err)
	}

	res, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	res.Body.Close()
	var body = make(map[string]string)
	err = json.Unmarshal(resBody, &body)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(res.Status)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)

	ref, err := url.Parse("upgrade")
	if err != nil {
		panic(err)
	}
	wsAddr := u.ResolveReference(ref)
	wsAddr.Scheme = "ws"
	values := wsAddr.Query()
	values.Add("t", body["ticket"])
	values.Add("id", body["userID"])
	wsAddr.RawQuery = values.Encode()

	conn, wres, err := websocket.DefaultDialer.DialContext(ctx, wsAddr.String(), http.Header{})
	if err != nil {
		panic(err)
	}
	return &ClientFields{
		Conn: conn,
		Res:  wres,
		Close: func() {
			if wres.StatusCode == 101 {
				conn.Close()
			}
			cancel()
		},
	}
}

func StartServerAndConnect(username string, password string, endpoint string, isIntTest ...bool) (*ServerFields, *ClientFields, func()) {
	utils.InitLogger(3)
	srv := StartServer(isIntTest...)
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
			Author:    user,
		})
	}
}
