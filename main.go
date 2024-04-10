package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	db, err := sql.Open("sqlite", "./medium.db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA locking_mode=IMMEDIATE;")
	db.Exec("pragma synchronous = normal;")
	db.Exec("pragma temp_store = memory;")
	// migrate(db)
	ctx := context.Background()
	queueChan := make(chan string)
	go func() {
		err = queryPages(ctx, db, queueChan)
		if err != nil {
			log.Panic("queue error", err)
		}
	}()
	startCount := 0
	endCount := 0
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		fmt.Printf("found %d new posts overall\n", endCount-startCount)
		db.Exec("vaccum;")
		db.Close()
		os.Exit(0)
	}()

	for {
		q := <-queueChan
		var next *Next
		for {
			parsed, newNext, err := importMedium(q, next)
			if err != nil {
				log.Panic("fetch error", err)
			}
			if len(parsed.collections) == 0 && len(parsed.tags) == 0 &&
				len(parsed.users) == 0 && len(parsed.posts) == 0 {
				break
			}
			// go func() {
			previousCount, err := countPosts(ctx, db)
			if err != nil {
				log.Panic("count error", err)
			}
			if startCount == 0 {
				startCount = previousCount
			}
			fmt.Printf("saving\n\t%d posts\n\t%d users\n\t%d collections\n\t%d tags\n",
				len(parsed.posts), len(parsed.users), len(parsed.collections), len(parsed.tags))
			err = save(ctx, db, parsed)
			if err != nil {
				log.Panic("save error", err)
			}
			err = logPage(ctx, db, q)
			if err != nil {
				log.Panic("log error", err)
			}
			endCount, err = countPosts(ctx, db)
			if err != nil {
				log.Panic("count 2 error", err)
			}
			fmt.Printf("found %d new posts\n", endCount-previousCount)
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
