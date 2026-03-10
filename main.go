package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/jim-ww/ytgo/internal/config"
	"github.com/jim-ww/ytgo/internal/player"
	"github.com/jim-ww/ytgo/internal/renderer"
	"github.com/jim-ww/ytgo/internal/rpc"
	"github.com/jim-ww/ytgo/internal/scraper"
	"github.com/jim-ww/ytgo/internal/store"
	"github.com/jim-ww/ytgo/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	client, err := rpc.NewClient(cfg.SocketPath, false)
	if err != nil {
		// first instance -> start server
		go func() {
			if err := rpc.StartServer(cfg.SocketPath); err != nil {
				log.Fatal(err)
			}
		}()
		client, err = rpc.NewClient(cfg.SocketPath, true) // connect
		if err != nil {
			log.Fatal(err)
		}
	}

	store, err := store.NewJSONStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	scraper := scraper.NewYouTubeSearcher()
	videoPlayer := player.NewMpvPlayer(cfg.TerminalVideo)
	imgRenderer := renderer.SixelRenderer{}

	p := tea.NewProgram(ui.NewRootModel(cfg, client, store, scraper, videoPlayer, imgRenderer))
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
