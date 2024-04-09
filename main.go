package main

import (
	"context"
	"database/sql"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "./medium.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA locking_mode=IMMEDIATE;")
	db.Exec("pragma synchronous = normal;	")
	db.Exec("pragma temp_store = memory;	")
	defer db.Close()
	// migrate(db)
	ctx := context.Background()
	queueChan := make(chan string)
	go func() {
		err = queryQueue(ctx, db, queueChan)
		if err != nil {
			log.Fatal("queue error", err)
		}
	}()

	for {
		q := <-queueChan
		err := importMedium(db, q, nil)
		if err != nil {
			log.Fatal("fetch error", err)
		}
	}

}
