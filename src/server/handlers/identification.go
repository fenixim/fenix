package handlers

import (
	"fenix/src/server"
	"fenix/src/websocket_models"
)

type IdentificationHandler struct {
	hub *server.ServerHub
}

func (i *IdentificationHandler) init() {
	i.hub.RegisterHandler("whoami", i.HandleWhoAmI)
}

func (i *IdentificationHandler) HandleWhoAmI(_ []byte, c *server.Client) {
	c.OutgoingPayloadQueue <- websocket_models.WhoAmI{
		ID:       c.User.UserID.Hex(),
		Username: c.User.Username,
	}
}

func NewIdentificationHandler(hub *server.ServerHub) {
	i := IdentificationHandler{hub: hub}
	i.init()
}