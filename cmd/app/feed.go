package main

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"time"

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
	GUID        string     `xml:"guid"`
	PublishedOn string     `xml:"pubDate"`
	Source      FeedSource `xml:"enclosure"`
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

	t, err = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", ep.PublishedOn)
	if err == nil {
		return t, nil
	}

	return t, nil
}

// duration gets en episode's time in seconds.
func (ep *FeedEpisode) duration() (float64, error) {
	t := 0.0

	resp, err := http.Get(ep.Source.URL)
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

// getEpisodes fetches the first 20 episodes from an XML feed.
func (app *application) getEpisodes(collectionID uint) ([]FeedEpisode, error) {
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

	return feed.Channel.Items[:20], nil
}

func (app *application) saveEpisodes(podcastID int, eps []FeedEpisode) error {
	for _, ep := range eps {
		pub, err := ep.publishedOnTime()
		if err != nil {
			return err
		}

		// TODO(charles): Figure out how to calculate duration for each
		//   episode in a goroutine so that this request can return in
		//   a timely fashion.
		// dur, err := ep.duration()
		// if err != nil {
		// 	return err
		// }

		err = app.episodes.Create(ep.Title, ep.GUID, podcastID, pub, 0)
		if err != nil {
			return err
		}
	}

	return nil
}
