package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

var startCount = 0

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := sql.Open("sqlite", "./medium.db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	// db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("pragma synchronous = normal;")
	db.Exec("pragma temp_store = memory;")
	db.Exec("pragma mmap_size = 30000000000;")
	ctx := context.Background()

	// pg, err := newDatabase(ctx, os.Getenv("PG_URL"))
	// if err != nil {
	// 	log.Panic(err)
	// }

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		endCount, _ := countPosts(ctx, db)
		fmt.Printf("found %d new posts overall\n", endCount-startCount)
		db.Close()
		err = exec.Command("sqlite3", "medium.db", "vacuum").Run()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	go func() {
		printStats(ctx, db)
		for {
			err = importMedium(ctx, db)
			if err != nil {
				log.Panic(err)
			}
			printStats(ctx, db)
		}
	}()

	// for {
	// 	var char rune
	// 	_, err := fmt.Scanf("%c", &char)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// }

	http.HandleFunc("/medium", api())
	http.Handle("/", http.FileServer(http.Dir("./web")))
	err = http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Panic(err)
	}

}

func generateLink(page Page) (string, error) {
	switch page.PageType {
	case 0:
		return "tags/" + page.ID, nil
	case 1:
		return "users/" + page.ID + "/profile", nil
	case 2:
		return "collections/" + page.ID, nil
	}
	return "", fmt.Errorf("unhandled page type: %d", page.PageType)
}

func importMedium(ctx context.Context, db *sql.DB) error {
	previousCount, err := countPosts(ctx, db)
	if err != nil {
		return fmt.Errorf("count error: %v", err)
	}
	if startCount == 0 {
		startCount = previousCount
	}
	pages, err := queryPages(ctx, db)
	if err != nil {
		log.Panic("queue error: ", err)
	}
	for _, page := range pages {
		link, err := generateLink(page)
		if err != nil {
			return fmt.Errorf("page error: %v", err)
		}
		var next *Next
		for {
			parsed, newNext, err := parseMedium(link, next)
			if err != nil {
				return fmt.Errorf("fetch error: %v", err)
			}
			if len(parsed.pages) == 0 && len(parsed.posts) == 0 {
				break
			}

			// fmt.Printf("\tsaving \t%d posts \t%d pages",
			// 	len(parsed.posts), len(parsed.pages))
			err = save(ctx, db, parsed)
			if err != nil {
				return fmt.Errorf("save error: %v", err)
			}

			// err = pg.save(ctx, parsed)
			// if err != nil {
			// 	log.Panic("save pg error", err)
			// }
			next = newNext
			if next == nil {
				break
			}
		}
		// fmt.Printf("\tfetched %d: %s", q.PageType, q.ID)
		err = logPage(ctx, db, page)
		if err != nil {
			return fmt.Errorf("log error: %v", err)
		}
	}
	endCount, err := countPosts(ctx, db)
	if err != nil {
		log.Panic("count 2 error", err)
	}
	fmt.Printf("found %d new posts\n", endCount-previousCount)

	return nil
}
