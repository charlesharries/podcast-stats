package main

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/tcolgate/mp3"
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
	PublishedOn string     `xml:"pubdate"`
	Source      FeedSource `xml:"enclosure"`
}

// FeedSource is a episode URL.
type FeedSource struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
}

// getEpisodes fetches the first 20 episodes from an XML feed.
func (app *application) getEpisodes(feedURL string) ([]FeedEpisode, error) {
	var feed FeedResults
	var blank []FeedEpisode

	// Request the data from the feed...
	resp, err := http.Get(feedURL)
	if err != nil {
		return blank, err
	}
	defer resp.Body.Close()

	// ... get the body...
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return blank, err
	}

	// ... unmarshal
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return blank, err
	}

	return feed.Channel.Items[:20], nil
}

func getEpisodeLength(URL string) (float64, error) {
	t := 0.0

	resp, err := http.Get(URL)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	d := mp3.NewDecoder(resp.Body)
	var f mp3.Frame
	skipped := 0

	for {
		err := d.Decode(&f, &skipped)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0.0, err
		}
		t = t + f.Duration().Seconds()
	}

	return t, nil
}
