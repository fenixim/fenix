package main

import (
	"fenix/src/database"
	"fenix/src/server/runner"
	"fenix/src/utils"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	wg := utils.NewWaitGroupCounter()
	env, err := godotenv.Read(".env")
	if err != nil {
		log.Panic("No .env file for database addresses!")
	}

	mongo_addr, ok := env["mongo_addr"]
	if !ok {
		log.Panic("Missing mongo_addr field in .env file")
	}

	db_name, ok := env["db_name"]
	if !ok {
		log.Panic("Missing db_name field in .env file")
	}

	hub := runner.NewHub(wg, database.NewMongoDatabase(mongo_addr, db_name))
	hub.Serve("0.0.0.0:8080")

	wg.Wait()
}
