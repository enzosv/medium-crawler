package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

type Post struct {
	title           string
	claps           int
	link            string
	publish_date    string
	collection      string
	recommend_count int
	response_count  int
	reading_time    float64
	tags            string
}

func main() {
	posts, err := query()
	if err != nil {
		log.Panic(err)
	}
	err = toCSV(posts)
	if err != nil {
		log.Panic(err)
	}
}

func toCSV(posts []Post) error {
	csvFile, err := os.Create("../web/medium.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()
	wr := csv.NewWriter(csvFile)
	for _, post := range posts {
		wr.Write([]string{
			strings.ReplaceAll(post.title, ",", "|"), fmt.Sprintf("%d", post.claps), post.link, post.publish_date, post.collection,
			fmt.Sprintf("%d", post.recommend_count), fmt.Sprintf("%d", post.response_count), fmt.Sprintf("%.2f", post.reading_time),
			post.tags,
		})
	}
	wr.Flush()
	return nil
}

func query() ([]Post, error) {
	db, err := sql.Open("sqlite", "../medium.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	query := `SELECT title, total_clap_count, 
    post_id, 
    date(published_at/1000, 'unixepoch'),
	COALESCE(c.name, ''), 
    recommend_count, response_count, reading_time, tags
    FROM posts p
    LEFT OUTER JOIN collections c
        ON c.collection_id = p.collection
    ORDER BY total_clap_count DESC
	;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.title, &post.claps, &post.link, &post.publish_date, &post.collection,
			&post.recommend_count, &post.response_count, &post.reading_time, &post.tags)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
