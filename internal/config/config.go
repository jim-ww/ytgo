package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
)

type Config struct {
	ConfigFile    kong.ConfigFlag `short:"c" name:"config" help:"Explicit path to config file" type:"path"`
	DataPath      string          `name:"data-path" default:"~/.local/share/ytgo/data.json" help:"Path to data.json (subscriptions, history, etc.)"`
	ThumbCacheDir string          `name:"thumb-dir" default:"/tmp/ytgo-thumbs" help:"Directory for cached thumbnails"`
	SocketPath    string          `name:"socket-path" default:"/tmp/ytgo.sock" help:"Unix socket for single-writer RPC"`
	TerminalVideo bool            `short:"t" name:"terminal-video" default:"false" help:"Use mpv -vo tct for in-terminal playback (experimental)"`
	SearchLimit   int             `name:"search-limit" default:"30" help:"Number of videos to fetch per request"`
}

// TODO: handle windows, macos users
func Load() (*Config, error) {
	var cfg Config

	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}

	paths := []string{
		filepath.Join(xdg, "ytgo", "config.toml"),
		filepath.Join(home, ".config", "ytgo", "config.toml"),
	}

	kong.Parse(&cfg,
		kong.Name("ytgo"),
		kong.Description("Minimalistic local-first YouTube TUI client"),
		kong.UsageOnError(),
		kong.Configuration(kongtoml.Loader, paths...),
	)

	cfg.DataPath = expandTilde(cfg.DataPath)
	cfg.ThumbCacheDir = expandTilde(cfg.ThumbCacheDir)
	cfg.SocketPath = expandTilde(cfg.SocketPath)

	return &cfg, nil
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}
