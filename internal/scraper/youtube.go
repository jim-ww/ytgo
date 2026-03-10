package scraper

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jim-ww/ytgo/internal/types"
)

type youtubeSearcher struct {
	client *http.Client
}

func NewYouTubeSearcher() YouTubeSearcher {
	return &youtubeSearcher{
		client: &http.Client{
			Timeout: 12 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

var ytInitialDataRegex = regexp.MustCompile(`(?s)var ytInitialData = (\{.*?\});`)

type ytInitialData struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						RichGridRenderer struct {
							Contents []json.RawMessage `json:"contents"`
						} `json:"richGridRenderer"`
					} `json:"content"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"twoColumnBrowseResultsRenderer"`
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents []struct {
						ItemSectionRenderer struct {
							Contents []json.RawMessage `json:"contents"`
						} `json:"itemSectionRenderer"`
					} `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
	} `json:"contents"`
}

type videoRenderer struct {
	VideoID string `json:"videoId"`
	Title   struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"title"`
	Thumbnail struct {
		Thumbnails []struct {
			URL string `json:"url"`
		} `json:"thumbnails"`
	} `json:"thumbnail"`
	LongBylineText struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"longBylineText"`
	LengthText struct {
		SimpleText string `json:"simpleText"`
	} `json:"lengthText"`
	ViewCountText struct {
		SimpleText string `json:"simpleText"`
	} `json:"viewCountText"`
	PublishedTimeText struct {
		SimpleText string `json:"simpleText"`
	} `json:"publishedTimeText"`
}

func (s *youtubeSearcher) Search(query string, opts SearchOptions) ([]types.Video, error) {
	if opts.Limit < 1 {
		opts.Limit = 10
	}
	if opts.SortBy == "" {
		opts.SortBy = SortRelevance
	}

	totalSteps := 2 + opts.Limit
	if opts.Progress != nil {
		opts.Progress(0, totalSteps)
	}

	isChannel := opts.Channel != ""
	pageURL := s.buildPageURL(query, opts, isChannel)

	videos, err := s.fetchAndParseVideos(pageURL, opts.Limit, opts.Progress, isChannel)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return nil, ErrNoVideosFound
	}

	if opts.Progress != nil {
		opts.Progress(2, totalSteps)
	}

	s.downloadThumbnailsConcurrently(videos, opts.Progress, totalSteps)
	s.client.Transport.(*http.Transport).CloseIdleConnections()

	return videos, nil
}

func (s *youtubeSearcher) buildPageURL(query string, opts SearchOptions, isChannel bool) string {
	if isChannel {
		chanID := normalizeChannelIdentifier(opts.Channel)
		if strings.HasPrefix(chanID, "UC") && len(chanID) == 24 {
			return fmt.Sprintf("https://www.youtube.com/channel/%s/videos", chanID)
		}
		return fmt.Sprintf("https://www.youtube.com/%s/videos", chanID)
	}

	sp := "EgIQAQ%253D%253D" // videos only
	switch opts.SortBy {
	case SortUploadDate:
		sp = "EgIIBA%253D%253D"
	case SortViewCount:
		sp = "EgIQBA%253D%253D"
	}

	return fmt.Sprintf("https://www.youtube.com/results?search_query=%s&sp=%s&hl=en&gl=US",
		url.QueryEscape(query), sp)
}

func normalizeChannelIdentifier(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "https://www.youtube.com/")
	input = strings.TrimPrefix(input, "http://www.youtube.com/")
	input = strings.TrimPrefix(input, "/")
	input = strings.TrimPrefix(input, "channel/")

	parts := strings.Split(input, "/")
	clean := parts[0]

	if strings.HasPrefix(clean, "@") {
		return clean
	}
	if strings.HasPrefix(clean, "UC") && len(clean) == 24 {
		return clean
	}
	return "@" + clean
}

func (s *youtubeSearcher) fetchAndParseVideos(pageURL string, limit int, progress func(int, int), isChannel bool) ([]types.Video, error) {
	resp, err := s.client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if progress != nil {
		progress(1, 2+limit)
	}

	return extractVideosFromBody(body, limit, isChannel)
}

func extractVideosFromBody(body []byte, limit int, isChannel bool) ([]types.Video, error) {
	matches := ytInitialDataRegex.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, ErrYtInitialDataNotFound
	}

	var initialData ytInitialData
	if err := json.Unmarshal(matches[1], &initialData); err != nil {
		return nil, fmt.Errorf("failed to parse ytInitialData: %w", err)
	}

	var containers []json.RawMessage
	if isChannel {
		for _, tab := range initialData.Contents.TwoColumnBrowseResultsRenderer.Tabs {
			if c := tab.TabRenderer.Content.RichGridRenderer.Contents; len(c) > 0 {
				containers = c
				break
			}
		}
	} else {
		if sections := initialData.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents; len(sections) > 0 {
			containers = sections[0].ItemSectionRenderer.Contents
		}
	}

	if len(containers) == 0 {
		return nil, ErrNoVideosFound
	}

	var videos []types.Video
	seen := make(map[string]bool)

	for _, raw := range containers {
		if len(videos) >= limit {
			break
		}

		if vr := extractVideoRenderer(raw); vr != nil {
			v := convertToVideo(vr)
			if v.URL != "" && !seen[v.URL] {
				seen[v.URL] = true
				videos = append(videos, v)
			}
		}
	}

	return videos, nil
}

func extractVideoRenderer(raw json.RawMessage) *videoRenderer {
	// Direct videoRenderer (search results)
	type direct struct {
		VideoRenderer *videoRenderer `json:"videoRenderer"`
	}
	var d direct
	if err := json.Unmarshal(raw, &d); err == nil && d.VideoRenderer != nil {
		return d.VideoRenderer
	}

	// richItemRenderer wrapper (current channel /videos tab)
	type wrapped struct {
		RichItemRenderer struct {
			Content struct {
				VideoRenderer *videoRenderer `json:"videoRenderer"`
			} `json:"content"`
		} `json:"richItemRenderer"`
	}
	var w wrapped
	if err := json.Unmarshal(raw, &w); err == nil && w.RichItemRenderer.Content.VideoRenderer != nil {
		return w.RichItemRenderer.Content.VideoRenderer
	}

	return nil
}

func convertToVideo(vr *videoRenderer) types.Video {
	var title, channel, thumb string
	if len(vr.Title.Runs) > 0 {
		title = vr.Title.Runs[0].Text
	}
	if len(vr.LongBylineText.Runs) > 0 {
		channel = vr.LongBylineText.Runs[0].Text
	}
	if len(vr.Thumbnail.Thumbnails) > 0 {
		thumb = vr.Thumbnail.Thumbnails[len(vr.Thumbnail.Thumbnails)-1].URL
	}

	return types.Video{
		Title:     title,
		URL:       "https://www.youtube.com/watch?v=" + vr.VideoID,
		Author:    channel,
		Duration:  vr.LengthText.SimpleText,
		Views:     vr.ViewCountText.SimpleText,
		Thumbnail: thumb,
		Published: vr.PublishedTimeText.SimpleText,
	}
}

func (s *youtubeSearcher) downloadThumbnailsConcurrently(videos []types.Video, progress func(int, int), totalSteps int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0

	cacheDir := "/tmp/ytgo-thumbs"
	_ = os.MkdirAll(cacheDir, 0o755)

	wg.Add(len(videos))
	for i := range videos {
		i := i
		go func() {
			defer wg.Done()

			path := downloadAndCacheThumbnail(s.client, videos[i].Thumbnail, cacheDir)
			if path == "" && videos[i].Thumbnail != "" {
				path = tryFallbackThumbnails(s.client, videos[i].Thumbnail, cacheDir)
			}

			mu.Lock()
			videos[i].ThumbnailPath = path
			completed++
			if progress != nil {
				progress(2+completed, totalSteps)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
}

func downloadAndCacheThumbnail(client *http.Client, thumbURL, cacheDir string) string {
	if thumbURL == "" {
		return ""
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(thumbURL)))
	path := filepath.Join(cacheDir, hash+".jpg")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	req, _ := http.NewRequest("GET", thumbURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.youtube.com/")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<19))
	if err != nil || len(data) == 0 || bytes.HasPrefix(data, []byte("<!DOCTYPE html")) {
		return ""
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return ""
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return ""
	}
	return path
}

func tryFallbackThumbnails(client *http.Client, original, cacheDir string) string {
	patterns := []string{"hqdefault", "mqdefault", "sddefault", "maxresdefault"}
	base := strings.TrimSuffix(original, "default.jpg")

	for _, qual := range patterns {
		candidate := base + qual + ".jpg"
		if candidate == original {
			continue
		}
		if path := downloadAndCacheThumbnail(client, candidate, cacheDir); path != "" {
			return path
		}
	}
	return ""
}
