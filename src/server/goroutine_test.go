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
}
