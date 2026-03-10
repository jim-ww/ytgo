package store

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jim-ww/ytgo/internal/config"
	"github.com/jim-ww/ytgo/internal/types"
)

const (
	CurrentVersion     = 1
	PlaylistLiked      = "liked"
	PlaylistDisliked   = "disliked"
	PlaylistWatchLater = "watchlater"
	PlaylistHistory    = "history"
)

type jsonStore struct {
	mu     sync.RWMutex
	cfg    *config.Config
	data   types.StoreData
	dirty  bool
	closed bool
}

func NewJSONStore(cfg *config.Config) (Store, error) {
	s := &jsonStore{
		cfg: cfg,
		data: types.StoreData{
			Version:       CurrentVersion,
			Videos:        make(map[string]types.Video),
			Playlists:     make(map[string]types.Playlist),
			Subscriptions: []types.Subscription{},
		},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	s.ensureSystemPlaylists()

	go s.autoSaveLoop()

	return s, nil
}

func (s *jsonStore) ensureSystemPlaylists() {
	s.mu.Lock()
	defer s.mu.Unlock()

	systems := []struct {
		id   string
		name string
	}{
		{PlaylistLiked, "Liked videos"},
		{PlaylistDisliked, "Disliked videos"},
		{PlaylistWatchLater, "Watch later"},
		{PlaylistHistory, "History"},
	}

	for _, sys := range systems {
		if _, exists := s.data.Playlists[sys.id]; !exists {
			s.data.Playlists[sys.id] = types.Playlist{
				ID:     sys.id,
				Name:   sys.name,
				Videos: []types.VideoRef{},
				System: true,
			}
			s.dirty = true
		}
	}
}

// Persistence

func (s *jsonStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := expandPath(s.cfg.DataPath)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&s.data); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	if s.data.Version != CurrentVersion {
		// In future: add migration code here
		return fmt.Errorf("unsupported store version: got %d, want %d", s.data.Version, CurrentVersion)
	}

	return nil
}

func (s *jsonStore) save() error {
	s.mu.RLock()
	if !s.dirty {
		s.mu.RUnlock()
		return nil
	}
	data := s.data // snapshot under read lock
	s.mu.RUnlock()

	path := expandPath(s.cfg.DataPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := f.Sync(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, path)
}

func (s *jsonStore) autoSaveLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		if s.closed {
			s.mu.RUnlock()
			return
		}
		s.mu.RUnlock()

		_ = s.Flush()
	}
}

func (s *jsonStore) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.dirty {
		return nil
	}
	err := s.save()
	if err == nil {
		s.dirty = false
	}
	return err
}

func (s *jsonStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return s.save()
}

// Read methods

func (s *jsonStore) GetAllVideos() map[string]types.Video {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]types.Video, len(s.data.Videos))
	maps.Copy(cp, s.data.Videos)
	return cp
}

func (s *jsonStore) GetVideo(id string) (types.Video, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data.Videos[id]
	return v, ok
}

func (s *jsonStore) GetPlaylist(id string) (types.Playlist, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.data.Playlists[id]
	return p, ok
}

func (s *jsonStore) GetSubscriptions() []types.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]types.Subscription, len(s.data.Subscriptions))
	copy(cp, s.data.Subscriptions)
	return cp
}

// Write methods

func (s *jsonStore) AddVideo(v types.Video) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v.ID == "" {
		return fmt.Errorf("video ID is required")
	}

	s.data.Videos[v.ID] = v
	s.dirty = true
	return nil
}

func (s *jsonStore) UpdateVideoProgress(id string, progress float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data.Videos[id]
	if !ok {
		return fmt.Errorf("video %s not found", id)
	}

	v.Progress = progress
	v.LastWatched = time.Now()
	s.data.Videos[id] = v
	s.dirty = true
	return nil
}

func (s *jsonStore) AddToSubscriptions(channelID, handle, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sub := range s.data.Subscriptions {
		if sub.ChannelID == channelID {
			return nil // already subscribed
		}
	}

	s.data.Subscriptions = append(s.data.Subscriptions, types.Subscription{
		ChannelID: channelID,
		Handle:    handle,
		Name:      name,
	})
	s.dirty = true
	return nil
}

func (s *jsonStore) RemoveFromSubscriptions(channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.data.Subscriptions {
		if sub.ChannelID == channelID {
			s.data.Subscriptions = append(s.data.Subscriptions[:i], s.data.Subscriptions[i+1:]...)
			s.dirty = true
			return nil
		}
	}
	return nil // not found
}

func (s *jsonStore) AddToPlaylist(playlistID, videoID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	playlist, ok := s.data.Playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist %s not found", playlistID)
	}

	// Check if video exists
	if _, exists := s.data.Videos[videoID]; !exists {
		return fmt.Errorf("video %s not in store", videoID)
	}

	// Avoid duplicates
	for _, ref := range playlist.Videos {
		if ref.ID == videoID {
			return nil
		}
	}

	playlist.Videos = append(playlist.Videos, types.VideoRef{
		ID:      videoID,
		AddedAt: time.Now(),
	})
	s.data.Playlists[playlistID] = playlist
	s.dirty = true
	return nil
}

func (s *jsonStore) RemoveFromPlaylist(playlistID, videoID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	playlist, ok := s.data.Playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist %s not found", playlistID)
	}

	for i, ref := range playlist.Videos {
		if ref.ID == videoID {
			playlist.Videos = append(playlist.Videos[:i], playlist.Videos[i+1:]...)
			s.data.Playlists[playlistID] = playlist
			s.dirty = true
			return nil
		}
	}
	return nil // not found
}

func (s *jsonStore) CreatePlaylist(name string, system bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("custom_%d", time.Now().UnixNano())

	_, exists := s.data.Playlists[id]
	if exists {
		return "", fmt.Errorf("id collision - very unlikely")
	}

	s.data.Playlists[id] = types.Playlist{
		ID:     id,
		Name:   name,
		Videos: []types.VideoRef{},
		System: system,
	}
	s.dirty = true
	return id, nil
}

func (s *jsonStore) DeletePlaylist(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	playlist, ok := s.data.Playlists[id]
	if !ok {
		return nil // not found
	}
	if playlist.System {
		return fmt.Errorf("cannot delete system playlist: %s", id)
	}

	delete(s.data.Playlists, id)
	s.dirty = true
	return nil
}

func expandPath(p string) string {
	if p == "" {
		return ""
	}
	if p[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[1:])
	}
	return filepath.Clean(p)
}
