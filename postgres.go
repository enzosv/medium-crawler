package main

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func newDatabase(ctx context.Context, url string) (Database, error) {
	var db Database
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return db, err
	}
	db.conn = conn
	return db, nil
}

func (db *Database) save(ctx context.Context, parsed Parsed) error {
	batch := &pgx.Batch{}
	if len(parsed.posts) > 0 {
		const query = `INSERT INTO posts(
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
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
		;`
		for _, post := range parsed.posts {
			var tags []string
			for _, tag := range post.Virtuals.Tags {
				tags = append(tags, tag.Slug)
			}
			batch.Queue(query,
				post.ID,
				post.Title,
				post.PublishedAt,
				post.UpdatedAt,
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
		}
	}
	if len(parsed.pages) > 0 {
		const query = `
		INSERT INTO pages(id, name, page_type) 
		values($1, $2, $3)
		ON CONFLICT (id, page_type) DO UPDATE SET 
			name = COALESCE(EXCLUDED.name, pages.name)
		`
		for _, page := range parsed.pages {
			batch.Queue(query, page.ID, page.Name, page.PageType)
		}
	}
	return db.conn.SendBatch(ctx, batch).Close()
}

func (db *Database) logPage(ctx context.Context, page Page) error {
	query := `UPDATE pages SET last_query = $1 WHERE id = $2 AND page_type = $3`
	_, err := db.conn.Exec(ctx, query, time.Now().Unix(), page.ID, page.PageType)
	return err
}
