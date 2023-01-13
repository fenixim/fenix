package testclient

import (
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"
)

func (m *TestClient) YodelCreate(t *testing.T, cli *test_utils.ClientFields, name string) {
	t.Helper()
	err := cli.Conn.WriteJSON(websocket_models.YodelCreate{Name: name}.SetType())
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func (m *TestClient) YodelGet(t *testing.T, cli *test_utils.ClientFields, yodelID string) {
	t.Helper()
	err := cli.Conn.WriteJSON(websocket_models.YodelGet{YodelID: yodelID}.SetType())
	if err != nil {
		t.Fatalf("%v", err)
	}
}
