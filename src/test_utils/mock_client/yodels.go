package mockclient

import (
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"
)

func (m *MockClient) YodelCreate(t *testing.T, cli *test_utils.ClientFields, name string) {
	t.Helper()
	err := cli.Conn.WriteJSON(websocket_models.YodelCreate{Name: name}.SetType())
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func (m *MockClient) YodelGet(t *testing.T, cli *test_utils.ClientFields, yodelID string) {
	t.Helper()
	err := cli.Conn.WriteJSON(websocket_models.YodelGet{YodelID: yodelID}.SetType())
	if err != nil {
		t.Fatalf("%v", err)
	}
}
