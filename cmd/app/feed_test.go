package main

import (
	"testing"
)

// TestGetEpisodes tests that we can actually fetch episodes.
func TestGetEpisodes(t *testing.T) {
	app := newTestApplication(t)
	collectionID := 941907967

	episodes, err := app.getEpisodes(collectionID)
	if err != nil {
		t.Fatal(err)
	}

	if len(episodes) != 20 {
		t.Errorf("want %d, got %d episodes", 20, len(episodes))
	}
}

// TestEpisodeSource tests that an episode has a source URL.
func TestEpisodeSource(t *testing.T) {
	app := newTestApplication(t)
	collectionID := 941907967

	episodes, err := app.getEpisodes(collectionID)
	if err != nil {
		t.Fatal(err)
	}

	if len(episodes[0].Source.URL) < 1 {
		t.Errorf("want feedURL to exist, got %q", episodes[0].Source.URL)
	}
}

// TestEpisodeLength tests that we can get the length of an individual episode.
func TestEpisodeLength(t *testing.T) {
	app := newTestApplication(t)
	collectionID := 941907967

	episodes, err := app.getEpisodes(collectionID)
	if err != nil {
		t.Fatal(err)
	}

	duration, err := episodes[0].duration()
	if err != nil {
		t.Fatal(err)
	}

	if duration == 0.0 {
		t.Errorf("want duration to be > 1, got %d (rounded)", duration)
	}
}
