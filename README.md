# What this does

- crawls medium using unofficial apis
- enables you to find the most popular posts

# How to

## I just want the list

1. No frontend yet. Go to something like https://inloop.github.io/sqlite-viewer/ and upload the [db](https://github.com/enzosv/medium-crawler/blob/main/medium.db)
2. Execute this query
   ```
   SELECT title, total_clap_count claps, 'https://medium.com/articles/' || post_id, date(published_at/1000, 'unixepoch') publish_date,
   c.name collection, recommend_count , response_count , reading_time ,tags
   FROM posts p
   LEFT OUTER JOIN collections c
   	ON c.collection_id = p.collection
   ORDER BY total_clap_count DESC;
   ```

## I want to update the list

1. `go run .`
