package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ROOTURL = "https://medium.com"

type Response struct {
	Payload struct {
		RelatedTags []Tag     `json:"relatedTags"`
		References  Reference `json:"references"`
		Paging      struct {
			Next *Next `json:"next"`
		} `json:"paging"`
	} `json:"payload"`
}

type Next struct {
	IgnoredIds []string `json:"ignoredIds"`
	To         string   `json:"to"`
	Page       int      `json:"page"`
}

type Reference struct {
	Post       map[string]Post       `json:"Post"`
	Collection map[string]Collection `json:"Collection"`
	User       map[string]User       `json:"User"`
}

type Tag struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type User struct {
	UserID string `json:"userId"`
}

type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Post struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PublishedAt int64  `json:"firstPublishedAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Collection  string `json:"homeCollectionId"`
	Creator     string `json:"creatorId"`
	IsPaid      bool   `json:"isMarkedPaywallOnly"`
	Virtuals    struct {
		ReadingTime    float64 `json:"readingTime"`
		TotalClapCount int     `json:"totalClapCount"`
		Tags           []Tag   `json:"tags"`
		Subtitle       string  `json:"subtitle"`
		RecommendCount int     `json:"recommends"`
		ResponseCount  int     `json:"responsesCreatedCount"`
	} `json:"virtuals"`
}

var lastRequest int64 = 0

// https://github.com/enzosv/easy-ios/blob/dcdcfbcf6333ecaae08cb0dfe7f940225fbdafa4/easy/Models/Resource.swift#L11
func fetchMedium(url string) (Response, error) {
	dif := time.Now().Unix() - lastRequest
	if dif < 3 {
		// avoid rate limit
		fmt.Println("sleeping", 3-dif)
		time.Sleep(time.Second * time.Duration(16-dif))
	}
	lastRequest = time.Now().Unix()
	fmt.Println("fetching", url)
	var response Response
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return response, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Referer", "https://twitter.com")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return response, err
	}
	res.Body.Close()
	data := strings.TrimPrefix(string(body), "])}while(1);</x>")

	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func importMedium(ctx context.Context, db *sql.DB, path string, next *Next) error {
	base := fmt.Sprintf("%s/_/api/%s/stream", ROOTURL, path)

	if next != nil {
		params := url.Values{}
		if len(next.IgnoredIds) > 0 {
			params.Add("ignoredIds", strings.Join(next.IgnoredIds, ","))
		}
		if next.Page > 1 {
			params.Add("page", fmt.Sprintf("%d", next.Page))
		}
		if next.To != "" {
			params.Add("to", next.To)
		}
		if len(params) > 0 {
			base += "?" + params.Encode()
		}
	}

	res, err := fetchMedium(base)
	if err != nil {
		return err
	}
	var users []User
	for _, user := range res.Payload.References.User {
		users = append(users, user)
	}
	var collections []Collection
	for _, collection := range res.Payload.References.Collection {
		collections = append(collections, collection)
	}
	tags := res.Payload.RelatedTags
	var posts []Post
	for _, post := range res.Payload.References.Post {
		posts = append(posts, post)
		users = append(users, User{post.Creator})
		if post.Collection != "" {
			collections = append(collections, Collection{post.Collection, ""})
		}
		for _, tag := range post.Virtuals.Tags {
			tags = append(tags, Tag{tag.Slug, tag.Name})
		}
	}
	err = save(ctx, db, res.Payload.RelatedTags, users, collections, posts)
	if err != nil {
		return err
	}
	newNext := res.Payload.Paging.Next
	if newNext != nil {
		if next != nil && (newNext.To == next.To || newNext.Page == next.Page) {
			// skip same next
			return nil
		}
		return importMedium(ctx, db, path, newNext)
	}
	return nil
}
