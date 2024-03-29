package mongo_interaction_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	"fenix/src/test_utils/test_client"
	"fenix/src/websocket_models"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestYodelIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("yodel creation results in new db entry", func(t *testing.T) {
		srv, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register", true)
		defer close()
		testClient := testclient.TestClient{}

		testClient.YodelCreate(t, cli, "Fenixland")

		var yodel websocket_models.Yodel
		err := cli.Conn.ReadJSON(&yodel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		yodel_ID, err := primitive.ObjectIDFromHex(yodel.YodelID)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		got := &database.Yodel{YodelID: yodel_ID}
		err = srv.Database.GetYodel(got)

		if err != nil {
			t.Fatalf("%q\n", err)
		}
		
		testClient.WhoAmI(t, cli)
		var whoAmI websocket_models.WhoAmI
		err = cli.Conn.ReadJSON(&whoAmI)
		if err != nil {
			t.Fatalf("%v\n", err)
		}


		expected := &database.Yodel{
			YodelID: yodel_ID,
			Name:    "Fenixland",
			Owner:   whoAmI.ID,
		}
		test_utils.AssertEqual(t, got, expected)
	})
}
