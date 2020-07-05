package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FeedResults is the full XML response.
type FeedResults struct {
	XMLName xml.Name    `xml:"rss"`
	Channel FeedChannel `xml:"channel"`
}

// FeedChannel is the channel belonging to the feed.
type FeedChannel struct {
	XMLName xml.Name      `xml:"channel"`
	Items   []FeedEpisode `xml:"item"`
}

// FeedEpisode is a single episode from the feed.
type FeedEpisode struct {
	XMLName     xml.Name   `xml:"item"`
	Title       string     `xml:"title"`
	GUID        string     `xml:"guid"`
	PublishedOn string     `xml:"pubDate"`
	Source      FeedSource `xml:"enclosure"`
	Duration    string     `xml:"duration"`
}

// FeedSource is a episode URL.
type FeedSource struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
}

// publishedOnTime gets a time.Time object for the episode's string time.
func (ep *FeedEpisode) publishedOnTime() (time.Time, error) {
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", ep.PublishedOn)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", ep.PublishedOn)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", ep.PublishedOn)
	if err == nil {
		return t, nil
	}

	return t, nil
}

// duration gets en episode's time in seconds.
func (ep *FeedEpisode) duration() (int, error) {
	// If no duration, just return 0
	if len(ep.Duration) == 0 {
		return 0, nil
	}

	// If duration doesn't have a :, it's likely already
	// in seconds.
	if !strings.Contains(ep.Duration, ":") {
		d, err := strconv.Atoi(ep.Duration)
		if err != nil {
			fmt.Println(ep.Title, ep.Duration)
			return 0, err
		}

		return d, nil
	}

	// Slice apart the duration on :, then add up the
	// seconds, minutes, and hours.
	parts := strings.Split(ep.Duration, ":")
	var s, m, h int
	var sec, min, hour string

	sec, parts = parts[len(parts)-1], parts[:len(parts)-1]
	s, err := strconv.Atoi(sec)
	if err != nil {
		return 0, err
	}

	if len(parts) > 0 {
		min, parts = parts[len(parts)-1], parts[:len(parts)-1]
		m, err = strconv.Atoi(min)
		if err != nil {
			return 0, err
		}
	}

	if len(parts) > 0 {
		hour, parts = parts[len(parts)-1], parts[:len(parts)-1]
		h, err = strconv.Atoi(hour)
		if err != nil {
			return 0, err
		}
	}

	return h*60*60 + m*60 + s, nil
}

// getEpisodes fetches the first 20 episodes from an XML feed.
func (app *application) getEpisodes(collectionID int) ([]FeedEpisode, error) {
	var feed FeedResults
	var blank []FeedEpisode

	podcast, err := app.podcasts.Find(collectionID)
	if err != nil {
		return blank, err
	}

	// Request the data from the feed...
	resp, err := http.Get(podcast.Feed)
	if err != nil {
		return blank, err
	}
	defer resp.Body.Close()

	// ... get the body...
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return blank, err
	}

	// ... and unmarshal.
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return blank, err
	}

	toGet := 20
	if len(feed.Channel.Items) < 20 {
		toGet = len(feed.Channel.Items)
	}

	return feed.Channel.Items[:toGet], nil
}

// saveEpisodes receives a list of episodes and saves them to the database.
func (app *application) saveEpisodes(podcastID int, eps []FeedEpisode) error {
	for _, ep := range eps {
		pub, err := ep.publishedOnTime()
		if err != nil {
			return err
		}

		dur, err := ep.duration()
		if err != nil {
			return err
		}

		err = app.episodes.Create(ep.Title, ep.GUID, ep.Source.URL, dur, podcastID, pub)
		if err != nil {
			return err
		}
	}

	return nil
}
