package server_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"
	"time"
)

func TestProtocols(t *testing.T) {
	t.Run("whoami", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		cli.Conn.WriteJSON(websocket_models.WhoAmI{}.SetType())

		var resProto websocket_models.WhoAmI
		cli.Conn.ReadJSON(&resProto)

		expected := "gopher123"
		got := resProto.Username
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("broadcast message", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		test_utils.MsgSend(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.MsgBroadcast
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Message
		expected := "General Kenobi, you are a bold one!"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("broadcast username", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		test_utils.MsgSend(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.MsgBroadcast
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Author.Username
		expected := "gopher123"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("send empty message", func(t *testing.T) {

		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		test_utils.MsgSend(t, cli, "")

		var resProto websocket_models.GenericError
		err := cli.Conn.ReadJSON(&resProto)
		if err != nil {
			t.Fatal(err)
		}

		expected := "message_empty"
		got := resProto.Error
		test_utils.AssertEqual(t, got, expected)
	})

	populate := func(srv *test_utils.ServerFields, count int) {
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

	t.Run("message history length", func(t *testing.T) {
		srv, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		populate(srv, 1)
		test_utils.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(test_utils.RecvMsgHistory(t, cli).Messages)
		expected := 1
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("empty message history length", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		test_utils.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(test_utils.RecvMsgHistory(t, cli).Messages)
		expected := 0
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("message history limit length", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()

		populate(srv, 51)
		test_utils.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(test_utils.RecvMsgHistory(t, cli).Messages)
		expected := 50
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("database error", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		srv.Database.ShouldErrorOnNext = true
		test_utils.MsgSend(t, cli, "this should error.")

		var resProto websocket_models.GenericError
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Error
		expected := "DatabaseError"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("server creation sends back id", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		test_utils.YodelCreate(t, cli)

		var yodel websocket_models.Yodel
		err := cli.Conn.ReadJSON(&yodel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := yodel.YodelID
		notExpected := ""
		test_utils.AssertNotEqual(t, got, notExpected)
	})
}
