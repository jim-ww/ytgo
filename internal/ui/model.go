package ui

import (
	"fmt"
	"image"
	"io"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/progress"
	"github.com/jim-ww/ytgo/internal/config"
	"github.com/jim-ww/ytgo/internal/player"
	"github.com/jim-ww/ytgo/internal/renderer"
	"github.com/jim-ww/ytgo/internal/rpc"
	"github.com/jim-ww/ytgo/internal/scraper"
	"github.com/jim-ww/ytgo/internal/store"
	"github.com/jim-ww/ytgo/internal/types"
)

type (
	searchResultsMsg  struct{ items []list.Item }
	statusMsg         string
	playbackStartMsg  struct{}
	progressUpdateMsg float64
	hideProgressMsg   struct{}
	progressTickMsg   struct{}
)

func progressTickCmd() tea.Cmd {
	return tea.Tick(time.Duration(rand.Int63n(int64(time.Second))), func(t time.Time) tea.Msg {
		return progressTickMsg{}
	})
}

var p *tea.Program

func SetProgram(pr *tea.Program) {
	p = pr
}

type viewMode string

const (
	modeSearch        viewMode = "search"
	modeSubscriptions viewMode = "subscriptions"
	modeHistory       viewMode = "history"
	modePlaylists     viewMode = "playlists"
	modeChannels      viewMode = "channels"
)

type Model struct {
	cfg             *config.Config
	client          *rpc.Client
	store           store.Store
	scraper         scraper.YouTubeSearcher
	player          player.Player
	renderer        renderer.ImageRenderer
	keyMap          keyMap
	help            help.Model
	searchInput     textinput.Model
	list            list.Model
	mode            viewMode
	searchFocused   bool
	loading         bool
	status          string
	width           int
	height          int
	showHelp        bool
	termPlayerMode  bool
	progress        progress.Model
	lastThumbRedraw time.Time
}

func NewRootModel(cfg *config.Config, client *rpc.Client, st store.Store, scr scraper.YouTubeSearcher, pl player.Player, rend renderer.ImageRenderer) *Model {
	m := &Model{
		cfg:            cfg,
		client:         client,
		store:          st,
		scraper:        scr,
		player:         pl,
		renderer:       rend,
		keyMap:         DefaultKeyMap,
		help:           help.New(),
		mode:           modeSearch, // modeSubscriptions,
		termPlayerMode: false,
		searchFocused:  true,
		progress:       progress.New(progress.WithoutPercentage(), progress.WithDefaultBlend()),
	}

	m.help.Styles = helpStyle

	ti := textinput.New()
	ti.Placeholder = "Search videos..."
	ti.CharLimit = 120
	ti.ShowSuggestions = true
	// TODO: store, display suggestions
	// ti.SetSuggestions([]string{""})
	ti.SetWidth(80)
	ti.KeyMap = DefaultKeyMap.searchInputMap
	m.searchInput = ti
	m.searchInput.Focus()

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItem
	delegate.Styles.SelectedDesc = selectedItem
	delegate.Styles.NormalTitle = normalItem
	delegate.Styles.NormalDesc = dimItem

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetShowTitle(false) // disable for now
	l.Styles.Title = listTitleStyle
	l.Title = string(m.mode)
	l.KeyMap = DefaultKeyMap.listKeyMap
	m.list = l

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCurrentMode(), textinput.Blink)
}

func (m *Model) send(msg tea.Msg) {
	if p != nil {
		p.Send(msg)
	}
}

func (m *Model) refreshCurrentMode() tea.Cmd {
	return func() tea.Msg {
		m.status = ""

		items := []list.Item{}

		// if query is not empty -> perform search, switch to searchMode
		if query := strings.TrimSpace(m.searchInput.Value()); query != "" {
			m.loading = true
			m.status = "Searching for: " + query
			m.mode = modeSearch
			m.list.Title = string(modeSearch)
			// TODO: imeplement fetching next 30 items, if on last page and pressed next
			opts := scraper.SearchOptions{Limit: 30, Progress: func(current, total int) {
				percent := float64(current) / float64(total)
				m.send(progressUpdateMsg(percent))
			}}
			videos, err := m.scraper.Search(query, opts)
			if err == nil {
				for i := range videos {
					items = append(items, videoItem{videos[i]})
				}
			}
		} else {
			switch m.mode {
			case modeSubscriptions:
			// fetch videos for all subscribed channels, via RSS to avoid rate limiting
			case modeChannels:
			// get all subscribed channels
			case modePlaylists:
			// display all local playlists, with preview set to first video thumbnail (if not empty)
			case modeHistory:
				// display watched videos history
			}
		}

		m.list.SetShowStatusBar(len(items) != 0)

		return searchResultsMsg{items: items}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	isInInputMode := m.searchFocused || m.list.FilterState() == list.Filtering

	resetStatusAfter := func(t time.Duration) tea.Cmd {
		return tea.Tick(t, func(time.Time) tea.Msg {
			return statusMsg("")
		})
	}
	// TODO: accept videoItem, channelItem, playlistItem
	redrawSelection := func() tea.Cmd {
		return func() tea.Msg {
			now := time.Now()
			if now.Sub(m.lastThumbRedraw) < 90*time.Millisecond {
				return nil
			}
			m.lastThumbRedraw = now

			sel := m.list.SelectedItem()
			if sel == nil {
				return nil
			}

			// right 33% of width, starting from ~30% height downward
			paneLeft := int(float64(m.width)*0.67) - 2
			paneTop := int(float64(m.height)*0.20) + 4

			// Available cells for thumbnail (tune these paddings!)
			maxCellsW := m.width - paneLeft - 3 // right margin
			maxCellsH := m.height - paneTop

			if maxCellsW <= 0 || maxCellsH <= 0 {
				return nil
			}

			vi, ok := sel.(videoItem)
			if !ok || vi.v.ThumbnailPath == "" || m.renderer == nil {
				_ = m.renderer.Clear(paneLeft, paneTop)
				return nil
			}

			// get real pixel dimensi8ns
			file, err := os.Open(vi.v.ThumbnailPath)
			if err != nil {
				return nil
			}
			defer file.Close()

			imgConfig, _, err := image.DecodeConfig(file)
			if err != nil {
				// fallback: 16:9
				imgConfig.Width = 1280
				imgConfig.Height = 720
			}
			// rewind file because DecodeConfig consumed it
			_, _ = file.Seek(0, io.SeekStart)

			origW := imgConfig.Width
			origH := imgConfig.Height
			if origW == 0 || origH == 0 {
				origW, origH = 1280, 720 // fallback
			}

			scale := math.Min(
				float64(maxCellsW)/float64(origW),
				float64(maxCellsH)/float64(origH),
			)

			drawW := int(math.Round(float64(origW) * scale))
			drawH := int(math.Round(float64(origH) * scale))

			const (
				pixelsPerCellX = 8
				pixelsPerCellY = 8
			)
			pixelW := drawW * pixelsPerCellX
			pixelH := drawH * pixelsPerCellY

			if pixelW > 1920 {
				pixelW = 1920
				pixelH = int(float64(pixelH) * 1920 / float64(pixelW))
			}
			if pixelH > 1080 {
				pixelH = 1080
				pixelW = int(float64(pixelW) * 1080 / float64(pixelH))
			}

			if drawW < 4 || drawH < 3 {
				_ = m.renderer.Clear(paneLeft, paneTop)
				return nil
			}

			offsetX := (maxCellsW - drawW) / 2
			offsetY := (maxCellsH - drawH) / 2

			drawX := paneLeft + offsetX
			drawY := paneTop + offsetY

			_ = m.renderer.Render(file, pixelW, pixelH, drawX, drawY)

			return nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// ignore input during tui video playing
		case m.termPlayerMode:
			break

		case key.Matches(msg, m.keyMap.Switch):
			m.searchFocused = !m.searchFocused
			if m.searchFocused {
				m.searchInput.Focus()
				cmds = append(cmds, textinput.Blink)
			} else {
				m.searchInput.Blur()
			}

		case key.Matches(msg, m.keyMap.Play):
			// Search mode
			if m.searchFocused && strings.TrimSpace(m.searchInput.Value()) != "" {
				m.searchFocused = false
				m.searchInput.Blur()
				cmds = append(cmds, m.refreshCurrentMode(), resetStatusAfter(time.Second))
			} else if m.list.FilterState() != list.Filtering {
				// List view mode
				// TODO: handle different types: videoItem, channelItem, playlistItem
				switch sel := m.list.SelectedItem().(type) {
				case videoItem:
					if m.player.IsAvailable() {
						m.status = " ▶ Playing: " + sel.v.Title + "..."
						m.loading = true
						cmds = append(cmds, resetStatusAfter(5*time.Second), func() tea.Msg {
							return playbackStartMsg{}
						}, progressTickCmd(), m.progress.SetPercent(0.0))
					}
					// should list videos for selected channel
					// case channelItem:

					// should list videos from selected playlist
					// case playlistItem:
				}
			}

		case key.Matches(msg, m.keyMap.listKeyMap.ClearFilter) && m.searchFocused:
			m.searchInput.Reset()

		// temporary fix for filtering breaking thumbnail preview
		case key.Matches(msg, m.keyMap.listKeyMap.Filter):
			// TODO: prevents first list item thumb render
			cmds = append(cmds, tea.ClearScreen)

		case isInInputMode:
			break

		case key.Matches(msg, m.keyMap.Quit):
			if m.client.IsOwner() {
				defer m.store.Flush()
			}
			cmds = append(cmds, tea.Quit)

		case key.Matches(msg, m.keyMap.listKeyMap.ShowFullHelp):
			m.showHelp = !m.showHelp
			m.help.ShowAll = m.showHelp

		case key.Matches(msg, m.keyMap.Refresh):
			cmds = append(cmds, tea.ClearScreen, redrawSelection())

		case key.Matches(msg, m.keyMap.listKeyMap.CursorUp, m.keyMap.listKeyMap.CursorDown, m.keyMap.listKeyMap.PrevPage, m.keyMap.listKeyMap.NextPage):
			cmds = append(cmds, redrawSelection())

			// disabled for now
			/*
				case key.Matches(msg, m.keyMap.Mode1):
					m.mode = modeSubscriptions
					m.list.Title = string(m.mode)
				case key.Matches(msg, m.keyMap.Mode2):
					m.mode = modePlaylists
					m.list.Title = string(m.mode)
				case key.Matches(msg, m.keyMap.Mode3):
					m.mode = modeHistory
					m.list.Title = string(m.mode)
				case key.Matches(msg, m.keyMap.Mode4):
					m.mode = modeChannels
					m.list.Title = string(m.mode)
			*/
		}

		cmds = append(cmds, cmd)

	case progressUpdateMsg:
		cmds = append(cmds, m.progress.SetPercent(float64(msg)))

	case progressTickMsg:
		if m.progress.Percent() < 1.0 {
			cmds = append(cmds, m.progress.IncrPercent(float64(rand.Intn(40))/100), progressTickCmd())
		} else {
			cmds = append(cmds, func() tea.Msg { return hideProgressMsg{} })
		}

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case playbackStartMsg:
		if sel, ok := m.list.SelectedItem().(videoItem); ok && m.player.IsAvailable() {
			cmd, err := m.player.Play(&sel.v, false)

			// if playing video in term, toggle alt mode, and wait for player to exit
			if m.cfg.TerminalVideo && err == nil && cmd != nil && cmd.Process != nil {
				cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
					m.termPlayerMode = true
					_ = cmd.Wait()
					m.termPlayerMode = false
					return tea.ClearScreen
				}))
			}
		}

	case searchResultsMsg:
		m.list.SetItems(msg.items)
		if len(msg.items) > 0 {
			m.list.Select(0)
		}
		cmds = append(cmds, redrawSelection(), m.progress.SetPercent(1.0), tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return hideProgressMsg{}
		}))

	case hideProgressMsg:
		m.loading = false
		cmds = append(cmds, m.progress.SetPercent(0.0))

	case statusMsg:
		m.status = string(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(m.width)
		m.searchInput.SetWidth(m.width - 4)
		m.progress.SetWidth(m.width)

		leftW := int(float64(m.width) * 0.65)
		m.list.SetSize(leftW-4, m.height-10)
		cmds = append(cmds, redrawSelection())

	}

	if m.searchFocused {
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	v := tea.NewView("")

	if m.termPlayerMode {
		v.SetContent("")
		v.AltScreen = false
		return v
	}

	if m.height < 12 || m.width < 60 {
		v.SetContent("Terminal too small")
		return v
	}

	searchLine := m.searchInput.View()
	if !m.searchFocused {
		if m.status != "" {
			searchLine = m.status
		} else {
			searchLine = "Press '" + m.keyMap.Switch.Keys()[0] + "' to switch focus" //"Mode: " + string(m.mode)
		}
	}
	searchBar := searchStyle.Width(m.width).Render(searchLine)

	progressBar := ""
	if m.loading {
		progressBar = progressStyle.Width(m.width).Render(m.progress.View())
	}

	leftW := int(float64(m.width) * 0.62) // 63-64
	// rightW := m.width - leftW - 3

	left := m.list.View()
	left = lipgloss.NewStyle().Width(leftW).Height(m.height - 10).PaddingTop(1).Render(left)

	// empty box
	right := "" // thumbnailStyle.Width(rightW).Height(m.height%37).Align(lipgloss.Center).Margin(13, 0, 13, 1).Render("Preview will be displayed here")

	split := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)

	helpView := m.help.View(m.keyMap)
	if !m.showHelp {
		helpView = m.help.ShortHelpView(m.keyMap.ShortHelp())
	}
	footer := footerStyle.Width(m.width).Render(helpView)

	content := lipgloss.JoinVertical(lipgloss.Top,
		searchBar,
		progressBar,
		split,
		footer,
	)

	v.SetContent(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// can either hold video / channel / playlist
type ListItem interface {
	Title() string
	Description() string
	FilterValue() string
	ThumbnailPath() string
}

type videoItem struct{ v types.Video }

func (i videoItem) Title() string { return i.v.Title }

func (i videoItem) Description() string {
	return fmt.Sprintf("%s • %s • %s • %s", i.v.Author, i.v.Duration, i.v.Published, i.v.Views)
}
func (i videoItem) FilterValue() string { return i.v.Title + " " + i.v.Author }

func (i videoItem) ThumbnailPath() string { return i.ThumbnailPath() }
