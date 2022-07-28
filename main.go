package main

import (
	"fenix/src/server"
	"fenix/src/utils"
)

func main() {
	// addr := flag.String("addr", ":8080", "http service address")
	// flag.Parse()
	wg := utils.WaitGroupCounter{}
	hub := server.NewHub(&wg)
	server.Serve("0.0.0.0:8080", &wg, hub)

	wg.Wait()
}
