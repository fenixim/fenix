package runner

import (
	"context"
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/server/handlers"
	"fenix/src/utils"
	"fenix/src/websocket_models"
	"sync"
)

func NewHub(wg *utils.WaitGroupCounter, database database.Database) *server.ServerHub {
	hub := server.ServerHub{
		Clients:           &sync.Map{},
		Broadcast_payload: make(chan websocket_models.JSONModel),
		Handlers:          make(map[string]func([]byte, *server.Client)),
		Wg:                wg,
		Database:          database,
	}

	handlers.NewMessageHandler(&hub)
	handlers.NewIdentificationHandler(&hub)
	handlers.NewYodelHandler(&hub)
	hub.Ctx, hub.Shutdown = context.WithCancel(context.Background())

	go hub.Run()

	return &hub
}
