package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
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
}

type User struct {
	UserID string  `json:"userId"`
	Name   *string `json:"name"`
}

type Collection struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

type Post struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PublishedAt int64  `json:"firstPublishedAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Collection  string `json:"homeCollectionId"`
	Creator     string `json:"creatorId"`
	IsPaid      bool   `json:"isSubscriptionLocked"`
	Virtuals    struct {
		ReadingTime    float64 `json:"readingTime"`
		TotalClapCount int     `json:"totalClapCount"`
		Tags           []Tag   `json:"tags"`
		Subtitle       string  `json:"subtitle"`
		RecommendCount int     `json:"recommends"`
		ResponseCount  int     `json:"responsesCreatedCount"`
	} `json:"virtuals"`
}

type Page struct {
	ID       string
	Name     *string
	PageType int
}

type Parsed struct {
	pages []Page
	posts []Post
}

const sleepDuration = 4

var lastRequest int64 = 0

// https://github.com/enzosv/easy-ios/blob/dcdcfbcf6333ecaae08cb0dfe7f940225fbdafa4/easy/Models/Resource.swift#L11
func fetchMedium(url string) (Response, error) {
	dif := time.Now().Unix() - lastRequest
	if dif < sleepDuration {
		// avoid rate limit
		fmt.Println("sleeping", sleepDuration-dif)
		time.Sleep(time.Second * time.Duration(sleepDuration-dif))
	}
	lastRequest = time.Now().Unix()
	fmt.Println("fetching", url)
	var response Response
	// out, err := normalFetch(url)
	out, err := curlFetch(url) // alternate between the two to avoid captcha
	if err != nil {
		return response, err
	}
	data := strings.TrimPrefix(string(out), "])}while(1);</x>")

	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return response, fmt.Errorf("json unmarshal %v", err)
	}
	return response, nil
}

func normalFetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("construct request %v", err)
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "sid=1:HD6zmkuwLRF1pGGvo4U5EEJGrnQOFTH/RnDEqD0cQEppJbTFTIyOfboKIOI1ha6c; uid=lo_7ae8ebac44cb")
	req.Header.Add("Host", "medium.com")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %v: %v", err, req)
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func curlFetch(url string) ([]byte, error) {
	return exec.Command("curl", url).Output()
}

func importMedium(path string, next *Next) (Parsed, *Next, error) {
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
		return Parsed{}, nil, err
	}
	var pages []Page

	for _, user := range res.Payload.References.User {
		pages = append(pages, Page{user.UserID, user.Name, 1})
	}
	for _, collection := range res.Payload.References.Collection {
		pages = append(pages, Page{collection.ID, collection.Name, 2})
	}
	for _, tag := range res.Payload.RelatedTags {
		pages = append(pages, Page{tag.Slug, nil, 0})
	}
	var posts []Post
	for _, post := range res.Payload.References.Post {
		posts = append(posts, post)
		pages = append(pages, Page{post.Creator, nil, 1})
		if post.Collection != "" {
			pages = append(pages, Page{post.Collection, nil, 2})
		}
		for _, tag := range post.Virtuals.Tags {
			pages = append(pages, Page{tag.Slug, nil, 0})
		}
	}
	parsed := Parsed{pages, posts}
	newNext := res.Payload.Paging.Next
	if newNext != nil {
		if next != nil && (newNext.To == next.To || newNext.Page == next.Page || reflect.DeepEqual(newNext.IgnoredIds, next.IgnoredIds)) {
			// skip same next
			return parsed, nil, nil
		}
		return parsed, newNext, nil
	}
	return parsed, nil, nil
}
