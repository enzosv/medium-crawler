package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func migrate(db *sql.DB) {
	migration := `
	CREATE TABLE tags (slug TEXT not null primary key, name TEXT not null);
	CREATE TABLE users (user_id TEXT not null primary key);
	CREATE TABLE collections (collection_id TEXT not null primary key, name TEXT not null);
	CREATE TABLE posts (
		post_id TEXT not null primary key,
		title TEXT not null,
		published_at INTEGER not null,
		updated_at INTEGER,
		collection TEXT,
		creator TEXT not null,
		is_paid INTEGER not null default 0,
		reading_time REAL,
		total_clap_count INTEGER,
		tags TEXT,
		subtitle TEXT,
		recommend_count INTEGER,
		response_count INTEGER
	)
	`
	_, err := db.Exec(migration)
	if err != nil {
		log.Printf("%q: %s\n", err, migration)
		return
	}

}

func save(ctx context.Context, db *sql.DB,
	_tags []Tag, _users []User, _collections []Collection, posts []Post,
) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if len(posts) > 0 {
		insert, err := tx.Prepare(`INSERT OR IGNORE INTO posts(
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
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`)
		if err != nil {
			return err
		}
		defer insert.Close()
		for _, post := range posts {
			var tags []string
			for _, tag := range post.Virtuals.Tags {
				tags = append(tags, tag.Slug)
				_tags = append(_tags, Tag{tag.Slug, tag.Name})
			}
			_users = append(_users, User{post.Creator})
			if post.Collection != "" {
				_collections = append(_collections, Collection{post.Collection, ""})
			}
			_, err = insert.ExecContext(ctx,
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
			if err != nil {
				return err
			}
		}
	}
	if len(_tags) > 0 {
		insert, err := tx.Prepare("INSERT OR IGNORE INTO tags(slug, name) values(?, ?)")
		if err != nil {
			return err
		}
		defer insert.Close()
		for _, tag := range _tags {
			_, err = insert.ExecContext(ctx, tag.Slug, tag.Name)
			if err != nil {
				return err
			}
		}
	}
	if len(_users) > 0 {
		insert, err := tx.Prepare("INSERT OR IGNORE INTO users(user_id) values(?)")
		if err != nil {
			return err
		}
		defer insert.Close()
		for _, user := range _users {
			_, err = insert.ExecContext(ctx, user.UserID)
			if err != nil {
				return err
			}
		}
	}
	if len(_collections) > 0 {
		insert, err := tx.Prepare("INSERT OR IGNORE INTO collections(collection_id, name) values(?, ?)")
		if err != nil {
			return err
		}
		defer insert.Close()
		for _, collection := range _collections {
			_, err = insert.ExecContext(ctx, collection.ID, collection.Name)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func queryTags(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "select slug from tags;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var slug string
		err = rows.Scan(&slug)
		if err != nil {
			log.Fatal(err)
		}
		tags = append(tags, slug)
	}
	return tags, rows.Close()
}

func queryIds(ctx context.Context, db *sql.DB, table, key string) ([]string, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("select %s from %s;", key, table))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Close()
}

type QueueItem struct {
	Type string
	ID   string
}

func queryQueue(ctx context.Context, db *sql.DB, idChan chan string) error {
	rows, err := db.QueryContext(ctx, `
	select creator, collection from posts
	order by  total_clap_count desc, recommend_count desc, response_count desc
	;`)
	if err != nil {
		log.Fatal(err)
	}
	queue := map[string]bool{}
	for rows.Next() {
		var creator string
		var collection *string
		err = rows.Scan(&creator, &collection)
		if err != nil {
			log.Fatal(err)
		}
		item := fmt.Sprintf("users/%s/profile", creator)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			idChan <- item
		}
		if collection != nil && *collection != "" {
			item = fmt.Sprintf("collections/%s", *collection)
			_, ok := queue[item]
			if !ok {
				queue[item] = true
				idChan <- item
			}
		}
	}
	rows.Close()
	// users
	rows, err = db.QueryContext(ctx, `select user_id from users;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		item := fmt.Sprintf("users/%s/profile", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			idChan <- item
		}
	}
	rows.Close()
	//collections
	rows, err = db.QueryContext(ctx, `select collection_id from collections;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		item := fmt.Sprintf("collections/%s", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			idChan <- item
		}
	}
	rows.Close()
	// tags
	rows, err = db.QueryContext(ctx, `select slug from tags;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		item := fmt.Sprintf("tags/%s", id)
		_, ok := queue[item]
		if !ok {
			queue[item] = true
			idChan <- item
		}
	}
	return rows.Close()
}
