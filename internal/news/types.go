package news

import "time"

// Article representa uma notícia extraída de um feed RSS.
type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Category    string    `json:"category"`
}

// CategoryInfo descreve uma categoria suportada e seus feeds.
type CategoryInfo struct {
	Key         string
	Title       string
	Emoji       string
	Description string
	FeedURLs    []FeedSource
}

// FeedSource define a fonte e URL de um feed RSS.
type FeedSource struct {
	Source string
	URL    string
}
