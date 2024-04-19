package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// TODO: parse https://topmediumstories.com/data/medium_1539563874.json

func countPosts(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	countQuery := db.QueryRowContext(ctx, "select count(*) from posts;")
	err := countQuery.Err()
	if err != nil {
		return count, err
	}
	err = countQuery.Scan(&count)
	return count, err
}

func save(ctx context.Context, db *sql.DB, parsed Parsed) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if len(parsed.posts) > 0 {
		insert, err := tx.Prepare(`INSERT INTO posts(
			post_id,
			title,
			published_at,
			updated_at,
			collection,
			creator,
			is_paid,
			reading_time,
			total_clap_count,
			tags,
			subtitle,
			recommend_count,
			response_count
		) values(
			?, ?, ?, ?, NULLIF(?,''), ?, ?, ?, ?, NULLIF(?,''), NULLIF(?,''), ?, ?
		)
		ON CONFLICT(post_id) 
  		DO UPDATE SET 
		  title = EXCLUDED.title,
		  published_at = EXCLUDED.published_at,
		  updated_at = EXCLUDED.updated_at,
		  collection = EXCLUDED.collection,
		  creator = EXCLUDED.creator,
		  is_paid = EXCLUDED.is_paid,
		  reading_time = EXCLUDED.reading_time,
		  total_clap_count = EXCLUDED.total_clap_count,
		  tags = EXCLUDED.tags,
		  subtitle = EXCLUDED.subtitle,
		  recommend_count = EXCLUDED.recommend_count,
		  response_count = EXCLUDED.response_count
		;`)
		if err != nil {
			return err
		}
		defer insert.Close()
		for _, post := range parsed.posts {
			var tags []string
			for _, tag := range post.Virtuals.Tags {
				tags = append(tags, tag.Slug)
			}
			_, err = insert.ExecContext(ctx,
				post.ID,
				post.Title,
				post.PublishedAt/1000,
				post.UpdatedAt/1000,
				post.Collection,
				post.Creator,
				post.IsPaid,
				post.Virtuals.ReadingTime,
				post.Virtuals.TotalClapCount,
				strings.Join(tags, ","),
				post.Virtuals.Subtitle,
				post.Virtuals.RecommendCount,
				post.Virtuals.ResponseCount,
			)
			if err != nil {
				return err
			}
		}
	}
	if len(parsed.pages) > 0 {
		logs, err := tx.Prepare(`
		INSERT INTO pages(id, name, page_type) 
		values(?, ?, ?)
		ON CONFLICT (id, page_type) DO UPDATE SET 
			name = COALESCE(EXCLUDED.name, pages.name)
		`)
		if err != nil {
			return err
		}
		defer logs.Close()
		for _, page := range parsed.pages {
			_, err = logs.ExecContext(ctx, page.ID, page.Name, page.PageType)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func queryIds(ctx context.Context, db *sql.DB, table, key string) ([]string, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s;", key, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Close()
}

func queryQueue(ctx context.Context, db *sql.DB, idChan chan string) error {
	rows, err := db.QueryContext(ctx, `
	SELECT creator, collection 
	FROM posts
	ORDER BY  total_clap_count desc
	;`)
	if err != nil {
		return err
	}
	defer rows.Close()
	queue := map[string]bool{}
	for rows.Next() {
		var creator string
		var collection *string
		err = rows.Scan(&creator, &collection)
		if err != nil {
			return err
		}
		item := fmt.Sprintf("users/%s/profile", creator)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			go func() { idChan <- item }()
		}
		if collection != nil && *collection != "" {
			item = fmt.Sprintf("collections/%s", *collection)
			_, ok := queue[item]
			if !ok {
				queue[item] = true
				go func() { idChan <- item }()
			}
		}
	}
	rows.Close()
	// users
	users, err := queryIds(ctx, db, "users", "user_id")
	if err != nil {
		return err
	}
	for _, id := range users {
		item := fmt.Sprintf("users/%s/profile", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			go func() { idChan <- item }()
		}
	}
	//collections
	collections, err := queryIds(ctx, db, "collections", "collection_id")
	if err != nil {
		return err
	}
	for _, id := range collections {
		item := fmt.Sprintf("collections/%s", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			go func() { idChan <- item }()
		}
	}
	// tags
	tags, err := queryIds(ctx, db, "tags", "slug")
	if err != nil {
		return err
	}
	for _, id := range tags {
		item := fmt.Sprintf("tags/%s", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			go func() { idChan <- item }()
		}
	}
	fmt.Println("queue", len(queue))
	return nil
}

func popularCollections(ctx context.Context, db *sql.DB) error {
	query := `SELECT c.name, SUM(p.total_clap_count) 
	FROM collections c
	LEFT OUTER JOIN posts p 
		ON p.collection = c.collection_id 
	GROUP BY c.collection_id 
	ORDER BY SUM(p.total_clap_count) DESC;`
	_, err := db.QueryContext(ctx, query)
	return err
}

func popularUsers(ctx context.Context, db *sql.DB) error {
	query := `SELECT u.user_id, SUM(p.total_clap_count)
	FROM users u
	LEFT OUTER JOIN posts p 
		ON p.creator = u.user_id  
	GROUP BY u.user_id 
	ORDER BY SUM(p.total_clap_count) DESC;`
	_, err := db.QueryContext(ctx, query)
	return err
}

func popularPosts(ctx context.Context, db *sql.DB) error {
	query := `SELECT title, total_clap_count claps, 'https://medium.com/articles/' || post_id, date(published_at/1000, 'unixepoch') publish_date, 
	c.name collection, recommend_count , response_count , reading_time ,tags 
	FROM posts p
	LEFT OUTER JOIN collections c
		ON c.collection_id = p.collection 
	ORDER BY total_clap_count DESC;`
	_, err := db.QueryContext(ctx, query)
	return err
}

func logPage(ctx context.Context, db *sql.DB, page Page) error {
	// query := `INSERT INTO pages (link, last_query)
	// values(?, ?)
	// ON CONFLICT(link)
	// DO UPDATE SET last_query = EXCLUDED.last_query`
	query := `UPDATE pages SET last_query = ? WHERE id = ? AND page_type = ?`
	_, err := db.ExecContext(ctx, query, time.Now().Unix(), page.ID, page.PageType)
	return err
}

func queryPages(ctx context.Context, db *sql.DB, idChan chan Page) error {
	rows, err := db.QueryContext(ctx, `
	SELECT id, page_type
	FROM pages
	ORDER BY last_query, page_type DESC
	;`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var page_type int

		err := rows.Scan(&id, &page_type)
		if err != nil {
			return err
		}
		idChan <- Page{id, nil, page_type}
	}
	return nil
}
