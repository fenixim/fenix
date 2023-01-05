package mongo_interaction_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	"fenix/src/websocket_models"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestYodelIntegration(t *testing.T) {
	t.Run("yodel creation results in new db entry", func(t *testing.T) {
		env, err := godotenv.Read("../../../.env")
		if err != nil {
			t.Fatalf("%v\n", err)
		}

		srv, cli, close := test_utils.StartServerAndConnect("gopher123",
			"mytotallyrealpassword", "/register", env)
		defer close()
		defer func() {
			
		}()

		test_utils.YodelCreate(t, cli, "Fenixland")

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
			Name: "Fenixland",
		}
		test_utils.AssertEqual(t, got, expected)
	})
}