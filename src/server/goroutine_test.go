package server_test

import (
	"fenix/src/test_utils"
	"testing"
	"time"
)

func TestEnsureGoroutinesStop(t *testing.T) {
	t.Run("when client exits", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()
		srv.Addr.Path = "/register"

		cli := test_utils.Connect("gopher123", "totallymypassword", srv.Addr)
		cli.Close()
		time.Sleep(10 * time.Millisecond)

		got := srv.Wg.Counter
		if got != 2 {
			srv.Wg.Names.Range(func(key, value interface{}) bool {
				t.Log(key)
				return true
			})
			t.Fail()
		}
	})
	t.Run("when server exits after client connected", func(t *testing.T) {

		srv := test_utils.StartServer()
		srv.Addr.Path = "/register"
		cli := test_utils.Connect("gopher123", "totallymypassword", srv.Addr)
		cli.Close()
		srv.Close()
		time.Sleep(5 * time.Second)

		srv.Wg.Wait()
		got := srv.Wg.Counter
		expected := 0
		if got != expected {
			t.Fail()
		}
	})
	t.Run("when server exits", func(t *testing.T) {

		srv := test_utils.StartServer()
		srv.Close()

		srv.Wg.Wait()
		got := srv.Wg.Counter
		expected := 0
		if got != expected {
			t.Fail()
		}
	})
}
