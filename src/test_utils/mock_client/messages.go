package mockclient

import (
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"
)

func (m *MockClient) MsgHistory(t *testing.T, cli *test_utils.ClientFields, from, to int64) {
	t.Helper()

	err := cli.Conn.WriteJSON(
		websocket_models.MsgHistory{From: from, To: to}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func (m *MockClient) MsgSend(t *testing.T, cli *test_utils.ClientFields, content string) {
	t.Helper()

	err := cli.Conn.WriteJSON(
		websocket_models.MsgSend{Message: content}.SetType())
	if err != nil {
		t.Fatal(err)
	}
}

func (m *MockClient) RecvMsgHistory(t *testing.T, cli *test_utils.ClientFields) websocket_models.MsgHistory {
	t.Helper()

	var resProto websocket_models.MsgHistory
	err := cli.Conn.ReadJSON(&resProto)
	if err != nil {
		t.Fatal(err)
	}

	return resProto
}
