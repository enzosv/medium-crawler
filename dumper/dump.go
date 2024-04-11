package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
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
	creator         string
	collection      string
	recommend_count int
	response_count  int
	reading_time    float64
	tags            string
	is_paid         int
}

func main() {
	posts, err := query()
	if err != nil {
		log.Panic(err)
	}
	err = toStringArray(posts)
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
		title := strings.ReplaceAll(post.title, ",", "|")
		title = strings.ReplaceAll(title, "\n", " ")
		wr.Write([]string{
			title, fmt.Sprintf("%d", post.claps), post.link, post.publish_date, post.collection,
			fmt.Sprintf("%d", post.recommend_count), fmt.Sprintf("%d", post.response_count), fmt.Sprintf("%.2f", post.reading_time),
			strings.ReplaceAll(post.tags, ",", "|"), fmt.Sprintf("%d", post.is_paid), post.creator,
		})
	}
	wr.Flush()
	return nil
}

func toStringArray(posts []Post) error {
	var lines [][]any
	for _, post := range posts {
		lines = append(lines, []any{
			post.title, post.claps, post.link, post.publish_date, post.collection,
			post.recommend_count, post.response_count, fmt.Sprintf("%.2f", post.reading_time),
			post.tags, post.is_paid, post.creator,
		})
	}
	file, err := json.Marshal(lines)
	if err != nil {
		return err
	}
	return os.WriteFile("../web/medium.json", file, 0644)
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
	COALESCE(u.name, ''),
	COALESCE(c.name, ''), 
    recommend_count, response_count, reading_time, tags, is_paid
    FROM posts p
    LEFT OUTER JOIN pages c
        ON c.id = p.collection
		AND c.page_type = 2
	LEFT OUTER JOIN pages u
		ON u.id = p.creator
		AND u.page_type = 1
	WHERE total_clap_count > 1000 OR published_at > date('now', '-1 month')
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
		err = rows.Scan(&post.title, &post.claps, &post.link, &post.publish_date, &post.creator, &post.collection,
			&post.recommend_count, &post.response_count, &post.reading_time, &post.tags, &post.is_paid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
