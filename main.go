package main

import (
	"context"
	"database/sql"
	"fmt"
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
		var next *Next
		for {
			parsed, newNext, err := importMedium(q, next)
			if err != nil {
				log.Fatal("fetch error", err)
			}
			go func() {
				err := save(ctx, db, parsed)
				if err != nil {
					log.Fatal("save error", err)
				}
				fmt.Printf("saved\n\t%d posts\n\t%d users\n\t%d collections\n\t%d tags\n",
					len(parsed.posts), len(parsed.users), len(parsed.collections), len(parsed.tags))
			}()
			next = newNext
			if next == nil {
				break
			}
		}
	}
}
