package main

import (
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/utils"

	"github.com/joho/godotenv"
)

func main() {
	wg := utils.WaitGroupCounter{}
	env, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}

	hub := server.NewHub(&wg, database.NewMongoDatabase(env["mongo_addr"], env["db_name"]))
	hub.Serve("0.0.0.0:8080")

	wg.Wait()
}
