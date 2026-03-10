<p align="center">
  <img src="./.assets/ytgo-logo.png" alt="ytgo logo" width="320" />
  <br/><br/>
  <strong>ytgo</strong> — minimalistic keyboard-first YouTube TUI client
  <br/>
</p>

<p align="center">
  <a href="https://github.com/jim-ww/ytgo/releases">
    <img src="https://img.shields.io/github/v/release/jim-ww/ytgo?color=green&logo=github" alt="Latest Release" />
  </a>
   <!-- <a href="https://github.com/jim-ww/ytgo">
     <img src="https://img.shields.io/github/stars/jim-ww/ytgo?style=social" alt="GitHub stars" />
   </a> -->
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/jim-ww/ytgo?color=blueviolet" alt="License: GPLv3" />
  </a>
  <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go version" />
  </a>
</p>

<p align="center">
  <img src=".assets/preview_v1.gif" alt="ytgo in action" width="720" />
  <br/>
  <em>Search • fuzzy filter • play • all from your terminal.</em>
</p>

**ytgo** is a fast, keyboard-driven terminal client for YouTube.

Search videos, fuzzy-filter results live, preview thumbnails and play directly in **mpv** — no Google API key/account needed.

## Installation

### Dependencies
**Requires:** [mpv](https://mpv.io/) installed on your system.

### Download pre-built binary (from Releases)

Go to the [Releases page](https://github.com/jim-ww/ytgo/releases) and download the archive matching your platform.

### With Go

```bash
go install github.com/jim-ww/ytgo@latest
```

### With Nix (via flake)

```bash
nix run github:jim-ww/ytgo
```
Or add to your flake inputs.

### From source

```bash
git clone https://github.com/jim-ww/ytgo
cd ytgo
go build -o ytgo
```

## Roadmap
- [ ] **proper infinite pagination** / load more
- [ ] **local management**: subscriptions, playlists, history, progress
- [ ] **client-server RPC (unix socket)** for safe concurrent file access from multiple instances
- [ ] **customization**: supply keybinding, styling colors via config
- [ ] **migration**: option to migrate data from Freetube, subscriptions from csv
- [ ] **alternative scraper backends**: e.g. Invidious

## Contributing
Bug reports, feature ideas, and pull requests are welcome.

## Donate
If ytgo saves you time or brings you joy, consider a small donation:

**Monero (XMR)**
`83YGRqP8uHed6NeegZQeX9ccCxbzoRHHEEi7pTwk4aqdJZEVXXA6NWtetnsEM2v33zFBBt3Rp6DNhU9qhJEGPspU14yN8t7`

All donations are greatly appreciated and go directly toward keeping the project alive and adding new features.

## License
**ytgo** is free software, licensed under the **GNU General Public License version 3** (GPLv3).

Everyone is free to view, copy, modify, distribute, and run the source code — with the condition that any derivative works are also distributed under the same GPLv3 license (or compatible terms).
This ensures the software remains free for all users forever.
