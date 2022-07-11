package main

import (
	"fenix/src/server"
	"fenix/src/utils"
	"flag"
)

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	wg := utils.WaitGroupCounter{}

	server.Serve(*addr, &wg)

	wg.Wait()
}
