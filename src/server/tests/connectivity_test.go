package server_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fenix/src/database"
	"fenix/src/test_utils"
	"net/http"
	"testing"
)

func TestStatusCodes(t *testing.T) {
	t.Run("login with no body errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/login"
		res, err := http.Post(srv.Addr.String(), "application/json", http.NoBody)
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("login with no username and password errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/login"

		b, err := json.Marshal(map[string]string{})
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("login for user that doesnt exist errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/login"

		b, err := json.Marshal(map[string]string{"username": "gopher123", "password": "pass"})
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("login for user with invalid password errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		srv.Addr.Path = "/login"

		b, err := json.Marshal(map[string]string{"username": "gopher123", "password": "notmypassword"})
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusForbidden

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("login for user with valid password is ok", func(t *testing.T) {
		srv := test_utils.StartServer()
		defer srv.Close()

		srv.Addr.Path = "/login"
		u := &database.User{
			Username: "gopher123",
			Password: []byte("pass"),
		}
		u.HashPassword()
		srv.Database.InsertUser(u)

		b, err := json.Marshal(map[string]string{"username": "gopher123", "password": "pass"})
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusOK

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("register with no body errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/register"
		res, err := http.Post(srv.Addr.String(), "application/json", http.NoBody)
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("register with no username and password errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/register"

		b, err := json.Marshal(map[string]string{})
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusBadRequest

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("register for user that exists errors", func(t *testing.T) {
		srv := test_utils.StartServer()
		u := &database.User{Username: "gopher123"}

		u.Salt = make([]byte, 16)

		rand.Read(u.Salt)
		u.Password = []byte("pass")
		u.HashPassword()

		err := srv.Database.InsertUser(u)
		if err != nil {
			t.Fatalf("%q\n", err)
		}

		srv.Addr.Path = "/register"
		b, err := json.Marshal(map[string]string{"username": "gopher123", "password": "pass"})
		if err != nil {
			t.Fatalf("%q\n", err)
		}
		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))

		if err != nil {
			t.Fatalf("%q\n", err)
		}
		
		got := res.StatusCode
		expected := http.StatusConflict

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("register for user that doesnt exist ok", func(t *testing.T) {
		srv := test_utils.StartServer()
		srv.Addr.Path = "/register"

		b, err := json.Marshal(map[string]string{"username": "gopher123", "password": "pass"})
		if err != nil {
			t.Fatalf("%q\n", err)
		}
		res, err := http.Post(srv.Addr.String(), "application/json", bytes.NewBuffer(b))

		if err != nil {
			t.Fatalf("%q\n", err)
		}

		got := res.StatusCode
		expected := http.StatusOK

		test_utils.AssertEqual(t, got, expected)
	})
}
