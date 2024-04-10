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
	defer db.Exec("vacuum;")
	defer db.Close()
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA locking_mode=IMMEDIATE;")
	db.Exec("pragma synchronous = normal;")
	db.Exec("pragma temp_store = memory;")
	// migrate(db)
	ctx := context.Background()
	queueChan := make(chan Page)
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
		db.Exec("vacuum;")
		db.Close()
		os.Exit(0)
	}()

	for {
		q := <-queueChan
		link := func() string {
			switch q.PageType {
			case 0:
				return "tags/" + q.ID
			case 1:
				return "users/" + q.ID + "/profile"
			case 2:
				return "collections/" + q.ID
			}
			log.Panic("unhandled page type", q.PageType)
			return ""
		}()
		var next *Next
		for {

			parsed, newNext, err := importMedium(link, next)
			if err != nil {
				log.Panic("fetch error", err)
			}
			if len(parsed.pages) == 0 && len(parsed.posts) == 0 {
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
			fmt.Printf("saving\n\t%d posts\n\t%d pages\n",
				len(parsed.posts), len(parsed.pages))
			err = save(ctx, db, parsed)
			if err != nil {
				log.Panic("save error", err)
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
		err = logPage(ctx, db, q)
		if err != nil {
			log.Panic("log error", err)
		}
	}
}
