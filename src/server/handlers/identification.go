package handlers

import (
	"encoding/json"
	"fenix/src/server"
	"fenix/src/websocket_models"
)

type IdentificationHandler struct {
	hub *server.ServerHub
}

func (i *IdentificationHandler) init() {
	i.hub.RegisterHandler("whoami", i.HandleWhoAmI)
}

func (i *IdentificationHandler) HandleWhoAmI(b []byte, c *server.Client) {
	whoami := &websocket_models.WhoAmI{}
	json.Unmarshal(b, whoami)
	c.OutgoingPayloadQueue <- websocket_models.WhoAmI{
		Nonce:    whoami.GetNonce(),
		ID:       c.User.UserID.Hex(),
		Username: c.User.Username,
	}
}

func NewIdentificationHandler(hub *server.ServerHub) {
	i := IdentificationHandler{hub: hub}
	i.init()
}
