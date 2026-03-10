package store

import "github.com/jim-ww/ytgo/internal/types"

type Store interface {
	// Read operations
	GetAllVideos() map[string]types.Video
	GetVideo(id string) (types.Video, bool)
	GetPlaylist(id string) (types.Playlist, bool)
	GetSubscriptions() []types.Subscription

	// Write operations
	AddVideo(v types.Video) error
	UpdateVideoProgress(id string, progress float64) error
	AddToSubscriptions(channelID, handle, name string) error
	RemoveFromSubscriptions(channelID string) error
	AddToPlaylist(playlistID, videoID string) error
	RemoveFromPlaylist(playlistID, videoID string) error
	CreatePlaylist(name string, system bool) (string, error)
	DeletePlaylist(id string) error

	// Persistence control
	Flush() error // force save now
	Close() error // called on shutdown
}
