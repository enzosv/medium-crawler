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
		err = queryPages(ctx, db, queueChan)
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
			if len(parsed.collections) == 0 && len(parsed.tags) == 0 &&
				len(parsed.users) == 0 && len(parsed.posts) == 0 {
				break
			}
			// go func() {
			previousCount, err := countPosts(ctx, db)
			if err != nil {
				log.Fatal("count error", err)
			}
			fmt.Printf("saving\n\t%d posts\n\t%d users\n\t%d collections\n\t%d tags\n",
				len(parsed.posts), len(parsed.users), len(parsed.collections), len(parsed.tags))
			err = save(ctx, db, parsed)
			if err != nil {
				log.Fatal("save error", err)
			}
			err = logPage(ctx, db, q)
			if err != nil {
				log.Fatal("log error", err)
			}
			newCount, err := countPosts(ctx, db)
			if err != nil {
				log.Fatal("count 2 error", err)
			}
			fmt.Printf("found %d new posts\n", newCount-previousCount)
			// if newCount == previousCount {
			// fast mode
			// break
			// }
			// }()
			next = newNext
			if next == nil {
				break
			}
		}
	}
}
