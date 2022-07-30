package main

import (
	"fenix/src/server"
	"fenix/src/utils"
)

func main() {
	wg := utils.WaitGroupCounter{}
	hub := server.NewHub(&wg)
	server.Serve("0.0.0.0:8080", &wg, hub)

	wg.Wait()
}
