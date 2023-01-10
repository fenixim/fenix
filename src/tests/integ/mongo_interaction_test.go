package mongo_interaction_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	mockclient "fenix/src/test_utils/mock_client"
	"fenix/src/websocket_models"
	"log"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestYodelIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("yodel creation results in new db entry", func(t *testing.T) {
		env, err := godotenv.Read("../../../.env")
		if err != nil {
			t.Fatalf("No .env file in project root %v", err)
		}
		_, ok := env["mongo_addr"]
		if !ok {
			t.Fatal("Missing mongo_addr field in .env file")
		}

		_, ok = env["integration_testing"]
		if !ok {
			log.Fatal("Missing integration_testing field in .env file")
		}

		srv, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register", env)
		defer close()
		mock := mockclient.MockClient{}
		mock.YodelCreate(t, cli, "Fenixland")

		var yodel websocket_models.Yodel
		err = cli.Conn.ReadJSON(&yodel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		yodel_ID, err := primitive.ObjectIDFromHex(yodel.YodelID)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		t.Logf("YodelID: %v\n", yodel_ID)
		got := &database.Yodel{YodelID: yodel_ID}
		err = srv.Database.GetYodel(got)

		if err != nil {
			t.Fatalf("%q\n", err)
		}

		expected := &database.Yodel{
			YodelID: yodel_ID,
			Name:    "Fenixland",
		}
		test_utils.AssertEqual(t, got, expected)
	})
}