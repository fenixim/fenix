package server_test

import (
	"fenix/src/test_utils"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

func TestStatusCodes(t *testing.T) {
	t.Run("no auth header", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		srv.Addr.Path = "/login"
		_, res, _ := websocket.DefaultDialer.Dial(srv.Addr.String(), http.Header{})

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user doesn't exist", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopherboi", "gopher1234", "/login")
		defer closeConn()

		got := cli.Res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)

	})

	t.Run("user does exist, wrong password", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		cli := test_utils.RegisterClient(t, srv, test_utils.Credentials{"gopher123", "gopher1234"})
		cli.Close()

		srv.Addr.Path = "/login"
		cli = test_utils.Connect("gopher123", "notmypassword", srv.Addr)

		got := cli.Res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user does exist, right password", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		test_utils.RegisterClient(t, srv, test_utils.Credentials{"gopher123", "gopher1234"})

		srv.Addr.Path = "/login"
		cli := test_utils.Connect("gopher123", "gopher1234", srv.Addr)

		got := cli.Res.StatusCode
		expected := http.StatusSwitchingProtocols

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("user already exists, register", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		test_utils.RegisterClient(t, srv, test_utils.Credentials{"gopher123", "gopher1234"})
		cli := test_utils.RegisterClient(t, srv, test_utils.Credentials{"gopher123", "gopher1234"})

		got := cli.Res.StatusCode
		expected := http.StatusConflict

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("for incorrect path", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "myawesomepassword", "/gophers")
		defer closeConn()

		got := cli.Res.StatusCode
		expected := http.StatusNotFound

		test_utils.AssertEqual(t, got, expected)
	})
}
