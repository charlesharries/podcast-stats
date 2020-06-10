package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// ITunesResult represents the whole response from the iTunes search API.
type ITunesResult struct {
	ResultsCount int
	Results      []Result
}

// Result corresponds to a single result from the iTunes search API.
type Result struct {
	CollectionID   int
	CollectionName string
	FeedURL        string
	ArtworkURL30   string
}

// getResults checks if a result exists in the redis cache. If it does, return
// it. Otherwise, fetch the results from iTunes instead.
func (app *application) getResults(term string) (ITunesResult, error) {
	var result ITunesResult

	// Check if there's an up-to-date result in the cache first.
	v, t, err := app.cache.Get(term)
	if err != nil {
		if !errors.Is(err, redis.ErrNil) {
			return result, err
		}
	}

	// If the result is less than 24 hours old, return the cache result.
	yesterday := time.Now().Add(-24 * time.Hour)
	if t > int(yesterday.Unix()) {
		err = json.Unmarshal([]byte(v), &result)
		if err != nil {
			return result, err
		}

		return result, nil
	}

	app.infoLog.Printf("cache miss: %q", term)

	// Make a request for our search results...
	req, err := http.NewRequest("GET", "https://itunes.apple.com/search", nil)
	if err != nil {
		return result, err
	}

	// ... add the querystring...
	q := req.URL.Query()
	q.Add("entity", "podcast")
	q.Add("term", term)
	req.URL.RawQuery = q.Encode()

	// ... create a client to make the request...
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}

	// ... read the response body...
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	// ... set the response in the cache...
	err = app.cache.Set(term, strings.TrimSpace(string(body)))
	if err != nil {
		return result, err
	}

	// ... unmarshal it into an ITunesResult struct...
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	// ... and return.
	return result, nil
}

// saveResults saves all podcasts in the results to the database.
func (app *application) saveResults(rs []Result) error {
	for _, r := range rs {
		err := app.podcasts.Create(r.CollectionID, r.CollectionName, r.FeedURL)
		if err != nil {
			return err
		}
	}

	return nil
}
