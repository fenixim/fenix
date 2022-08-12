package server

import (
	"testing"
	"time"
)

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
		time.Sleep(5 * time.Second)

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
}
