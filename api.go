package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/amalfra/etag"
	_ "modernc.org/sqlite"
)

type SimplePost struct {
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

func api() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		posts, err := query()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
			return
		}
		lines := toArray(posts)
		w.Header().Set("Content-Type", "application/json")
		body, err := json.Marshal(lines)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
			return
		}
		// w.Header().Set("Cache-Control", "max-age=600")
		w.Header().Set("etag", etag.Generate(string(body), true))
		w.Write(body)
	}
}

func toArray(posts []SimplePost) [][]any {
	var lines [][]any
	for _, post := range posts {
		lines = append(lines, []any{
			post.title, post.claps, post.link, post.publish_date, post.collection,
			post.recommend_count, post.response_count, math.Round(post.reading_time),
			post.tags, post.is_paid, post.creator,
		})
	}
	return lines
}

func query() ([]SimplePost, error) {
	db, err := sql.Open("sqlite", "./medium.db?mode=ro&_journal=WAL")
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
	var posts []SimplePost
	for rows.Next() {
		var post SimplePost
		err = rows.Scan(&post.title, &post.claps, &post.link, &post.publish_date, &post.creator, &post.collection,
			&post.recommend_count, &post.response_count, &post.reading_time, &post.tags, &post.is_paid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	fmt.Println("done query")
	return posts, nil
}
