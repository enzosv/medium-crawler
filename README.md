# What this does

- crawls medium using unofficial apis
- enables you to find the most popular posts

# How to

## I just want the list

- Go to https://enzosv.github.io/medium-crawler/

## I want to analyze the list

- `sqlite3 medium.db`

## I want to update the list

### Requirements

- [curl impersonate](https://github.com/lwthiker/curl-impersonate)
- go v1.21.0+
  - untested on older versions

### Run

- `go run .`

# Notes

- Some of the posts might be responses. I checked too late.
- This probably breaks medium TOS
- The list/db isn't always updated and is far from comprehensive
