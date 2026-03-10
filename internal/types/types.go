package types

import "time"

type Video struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Thumbnail     string    `json:"thumbnail"`
	Duration      string    `json:"duration"`
	Progress      float64   `json:"progress"` // 0.0 - 1.0
	Views         string    `json:"views"`
	LastWatched   time.Time `json:"lastWatched"`
	URL           string    `json:"url"`
	ThumbnailPath string    `json:"thumbnail_path"` // local path for preview
	Description   string    `json:"-"`
	Published     string    `json:"published"` // relative published/upload date
}

type VideoRef struct {
	ID      string    `json:"id"`
	AddedAt time.Time `json:"addedAt"`
}

type Playlist struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Videos []VideoRef `json:"videos"`
	System bool       `json:"system"` // liked, history, watchlater, disliked
}

type Subscription struct {
	ChannelID string `json:"channelId"`
	Handle    string `json:"handle"`
	Name      string `json:"name"`
}

type StoreData struct {
	Version       int                 `json:"version"` // start at 1
	Videos        map[string]Video    `json:"videos"`  // central deduped store
	Playlists     map[string]Playlist `json:"playlists"`
	Subscriptions []Subscription      `json:"subscriptions"`
}
