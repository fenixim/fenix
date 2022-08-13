package server_test

import (
	"fenix/src/test_utils"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

type credentials struct {
	username string
	password string
}

func register(t *testing.T, srv *serverFields, auth credentials) *clientFields {
	t.Helper()

	srv.addr.Path = "/register"
	cli := Connect_(auth.username, auth.password, srv.addr)

	return cli
}

func TestStatusCodes(t *testing.T) {
	t.Run("no auth header", func(t *testing.T) {
		srv := StartServer_()
		defer srv.close()

		srv.addr.Path = "/login"
		_, res, _ := websocket.DefaultDialer.Dial(srv.addr.String(), http.Header{})

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user doesn't exist", func(t *testing.T) {
		_, cli, closeConn := StartServerAndConnect("gopherboi", "gopher1234", "/login")
		defer closeConn()

		got := cli.res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)

	})

	t.Run("user does exist, wrong password", func(t *testing.T) {
		srv := StartServer_()
		defer srv.close()

		register(t, srv, credentials{"gopher123", "gopher1234"})

		srv.addr.Path = "/login"
		cli := Connect_("gopher123", "notmypassword", srv.addr)

		got := cli.res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user does exist, right password", func(t *testing.T) {
		srv := StartServer_()
		defer srv.close()

		register(t, srv, credentials{"gopher123", "gopher1234"})

		srv.addr.Path = "/login"
		cli := Connect_("gopher123", "gopher1234", srv.addr)

		got := cli.res.StatusCode
		expected := http.StatusSwitchingProtocols

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user already exists, register", func(t *testing.T) {
		srv := StartServer_()
		defer srv.close()

		register(t, srv, credentials{"gopher123", "gopher1234"})
		cli := register(t, srv, credentials{"gopher123", "gopher1234"})

		got := cli.res.StatusCode
		expected := http.StatusConflict

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("for incorrect path", func(t *testing.T) {
		_, cli, closeConn := StartServerAndConnect("gopher123", "myawesomepassword", "/gophers")
		defer closeConn()

		got := cli.res.StatusCode
		expected := http.StatusNotFound

		test_utils.AssertEqual(t, got, expected)
	})
}
