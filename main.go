package main

import (
	"fenix/src/database"
	"fenix/src/server"
	"fenix/src/utils"
	"log"
	"os"
)

func getMongoDB() database.Database {
	var db database.Database
	mongoAddr := os.Getenv("mongo_addr")
	dbName := os.Getenv("database")

	if mongoAddr == "" || dbName == "" {
		log.Panicf("Couldn't get database env -  mongoAddr: %q   intTest: %q", mongoAddr, dbName)
	} else {
		db = database.NewMongoDatabase(mongoAddr, dbName)
		err := db.ClearDB()
		if err != nil {
			panic(err)
		}
	}
	return db
}

func main() {
	wg := utils.NewWaitGroupCounter()

	hub := server.NewHub(wg, getMongoDB())
	hub.Serve("0.0.0.0:8080")

	wg.Wait()
}
