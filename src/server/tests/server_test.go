package server_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	mockclient "fenix/src/test_utils/mock_client"
	"fenix/src/websocket_models"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWhoAmIHandlers(t *testing.T) {
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
}

func TestMessageHandlers(t *testing.T) {
	t.Run("msg_broadcast has correct content", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		mock := mockclient.MockClient{}
		mock.MsgSend(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.MsgBroadcast
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Message
		expected := "General Kenobi, you are a bold one!"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("msg_broadcast has correct username", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		mock := mockclient.MockClient{}
		mock.MsgSend(t, cli, "General Kenobi, you are a bold one!")

		var resProto websocket_models.MsgBroadcast
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Author.Username
		expected := "gopher123"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("msg_broadcast doesnt allow empty messages", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		mock := mockclient.MockClient{}
		mock.MsgSend(t, cli, "")

		var resProto websocket_models.GenericError
		err := cli.Conn.ReadJSON(&resProto)
		if err != nil {
			t.Fatal(err)
		}

		expected := "message_empty"
		got := resProto.Error
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("msg_history has correct length", func(t *testing.T) {
		srv, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		test_utils.Populate(srv, 1)
		mock := mockclient.MockClient{}
		mock.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(mock.RecvMsgHistory(t, cli).Messages)
		expected := 1
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("empty message history length", func(t *testing.T) {
		_, cli, closeConn := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer closeConn()

		mock := mockclient.MockClient{}
		mock.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(mock.RecvMsgHistory(t, cli).Messages)
		expected := 0
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("message history limit length", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123", "mytotallyrealpassword", "/register")
		defer close()

		test_utils.Populate(srv, 51)
		mock := mockclient.MockClient{}
		mock.MsgHistory(t, cli, 0, time.Now().UnixNano())

		got := len(mock.RecvMsgHistory(t, cli).Messages)
		expected := 50
		test_utils.AssertEqual(t, got, expected)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("database error", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123", "pass", "/register")
		defer close()

		srv.Database.(*database.InMemoryDatabase).ShouldErrorOnNext = true
		mock := mockclient.MockClient{}
		mock.MsgSend(t, cli, "this should error.")

		var resProto websocket_models.GenericError
		cli.Conn.ReadJSON(&resProto)

		got := resProto.Error
		expected := "DatabaseError"
		test_utils.AssertEqual(t, got, expected)
	})
}

func TestYodelHandlers(t *testing.T) {
	t.Run("yodel creation results in valid id", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		mock := mockclient.MockClient{}
		mock.YodelCreate(t, cli, "Fenixland")

		var yodel websocket_models.Yodel
		err := cli.Conn.ReadJSON(&yodel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := yodel.YodelID
		if !primitive.IsValidObjectID(got) {
			t.Errorf("Invalid YodelID: %q", got)
		}
	})

	t.Run("yodel creation sends back name", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		mock := mockclient.MockClient{}
		mock.YodelCreate(t, cli, "Fenixland")

		var yodel websocket_models.Yodel
		err := cli.Conn.ReadJSON(&yodel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := yodel.Name
		expected := "Fenixland"
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("users can request yodel info", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		yodel := &database.Yodel{Name: "Yodelyay"}
		srv.Database.InsertYodel(yodel)

		mock := mockclient.MockClient{}
		mock.YodelGet(t, cli, yodel.YodelID.Hex())

		var res websocket_models.Yodel
		err := cli.Conn.ReadJSON(&res)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		resYodelID, err := primitive.ObjectIDFromHex(res.YodelID)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := &database.Yodel{YodelID: resYodelID, Name: res.Name}
		expected := yodel

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("users requesting yodel info with invalid hex errors", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		mock := mockclient.MockClient{}
		mock.YodelGet(t, cli, "z0")

		var res websocket_models.GenericError
		err := cli.Conn.ReadJSON(&res)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := res.T
		expected := websocket_models.GenericError{}.Type()

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("ensure when user request yodel that does not exist errors", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		mock := mockclient.MockClient{}
		mock.YodelGet(t, cli, primitive.NewObjectIDFromTimestamp(time.Now()).Hex())

		var res websocket_models.GenericError
		err := cli.Conn.ReadJSON(&res)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := res.T
		expected := websocket_models.GenericError{}.Type()

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("users requesting yodel with no ID errors", func(t *testing.T) {
		_, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register")
		defer close()

		err := cli.Conn.WriteJSON(websocket_models.YodelGet{}.SetType())
		if err != nil {
			t.Fatalf("%v\n", err)

		}

		var res websocket_models.GenericError
		err = cli.Conn.ReadJSON(&res)
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		got := res.T
		expected := websocket_models.GenericError{}.Type()

		test_utils.AssertEqual(t, got, expected)
	})
}
