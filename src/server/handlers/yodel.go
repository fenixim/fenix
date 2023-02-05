package handlers

import (
	"encoding/json"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/utils"
	"fenix/src/websocket_models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type YodelHandler struct {
	hub *server.ServerHub
}

func (y *YodelHandler) init() {
	y.hub.RegisterHandler(websocket_models.YodelCreate{}.Type(), y.HandleYodelCreate)
	y.hub.RegisterHandler(websocket_models.YodelGet{}.Type(), y.HandleYodelGet)
}

func (y *YodelHandler) HandleYodelCreate(b []byte, c *server.Client) {
	var yodel websocket_models.YodelCreate
	err := json.Unmarshal(b, &yodel)
	if err != nil {
		utils.InfoLogger.Printf("error in decoding yodelcreate json: %v", err)
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "JSONDecodeError"}
		return
	}

	if yodel.Name == "" {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{
			Error:   "YodelNameEmpty",
			Message: "Cannot create a yodel with no name!",
		}
		return
	}

	db_yodel := &database.Yodel{
		Name:  yodel.Name,
		Owner: c.User.UserID.Hex(),
	}

	err = y.hub.Database.InsertYodel(db_yodel)

	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "DatabaseError"}
		return
	}

	c.OutgoingPayloadQueue <- websocket_models.Yodel{
		YodelID: db_yodel.YodelID.Hex(),
		Name:    yodel.Name,
		Owner:   c.User.UserID.Hex(),
	}
}

func (y *YodelHandler) HandleYodelGet(b []byte, c *server.Client) {
	var yodelGet websocket_models.YodelGet
	err := json.Unmarshal(b, &yodelGet)
	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "JSONDecodeError"}
		utils.InfoLogger.Printf("error in decoding yodelcreate json: %q\n", err)
		return
	}
	if yodelGet.YodelID == "" {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "MissingIDError", Message: "ID field cannot be empty!"}
		return
	}
	yodelID, err := primitive.ObjectIDFromHex(yodelGet.YodelID)
	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "IDFormattingError", Message: "ID field is formatted incorrectly!"}
		return
	}

	yodel := database.Yodel{YodelID: yodelID}
	err = y.hub.Database.GetYodel(&yodel)
	if err != nil {
		c.OutgoingPayloadQueue <- websocket_models.GenericError{Error: "YodelDoesntExistError"}
		return
	}

	c.OutgoingPayloadQueue <- websocket_models.Yodel{
		YodelID: yodel.YodelID.Hex(),
		Name:    yodel.Name,
	}
}

func NewYodelHandler(hub *server.ServerHub) *YodelHandler {
	y := YodelHandler{hub: hub}
	y.init()
	return &y
}
