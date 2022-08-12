package server

import (
	"crypto/rand"
	"fenix/src/database"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

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
