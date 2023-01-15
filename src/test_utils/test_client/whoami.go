package testclient

import (
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"
)

func (m *TestClient) WhoAmI(t *testing.T, cli *test_utils.ClientFields) {
	err := cli.Conn.WriteJSON(websocket_models.WhoAmI{}.SetType())
	if err != nil {
		t.Fatalf("%q\n", err)
	}
}