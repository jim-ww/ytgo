package scraper

import (
	"errors"

	"github.com/jim-ww/ytgo/internal/types"
)

var (
	ErrYtInitialDataNotFound = errors.New("ytInitialData script not found in response")
	ErrNoVideosFound         = errors.New("no videos extracted from response")
)

type VideoSort string

const (
	SortRelevance  VideoSort = "relevance"
	SortUploadDate VideoSort = "upload_date"
	SortViewCount  VideoSort = "view_count"
)

type SearchOptions struct {
	Limit    int
	Channel  string // @handle, UCxxxx, full URL, or just the name
	SortBy   VideoSort
	Progress func(current, total int)
}

type YouTubeSearcher interface {
	Search(query string, opts SearchOptions) ([]types.Video, error)
}
