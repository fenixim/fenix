package main

import (
	"fenix/src/database"
	"fenix/src/server/runner"
	"fenix/src/utils"
	"log"
	"os"
	"strconv"
)

func getMongoDB() database.Database {
	var db database.Database
	mongoAddr := os.Getenv("mongo_addr")
	dbName := os.Getenv("db_name")

	if mongoAddr == "" || dbName == "" {
		log.Panicf("Couldn't get database env -  mongoAddr: %q   dbName: %q", mongoAddr, dbName)
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
	level := os.Getenv("log_level")
	i, err := strconv.ParseInt(level, 10, 8);
	if err != nil {
		panic(err)
	}
	
	utils.InitLogger(utils.LogLevel(i), "main.log")
	hub := runner.NewHub(wg, getMongoDB())
	hub.Serve("0.0.0.0:8080")

	wg.Wait()
}
