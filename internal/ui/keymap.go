package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
)

type keyMap struct {
	Quit           key.Binding
	Refresh        key.Binding
	Play           key.Binding
	Delete         key.Binding
	Switch         key.Binding
	Mode1          key.Binding
	Mode2          key.Binding
	Mode3          key.Binding
	Mode4          key.Binding
	listKeyMap     list.KeyMap
	searchInputMap textinput.KeyMap
}

var DefaultKeyMap = keyMap{
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Play:       key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter", "play")),
	Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Switch:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
	Mode1:      key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "subscriptions")),
	Mode2:      key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "history")),
	Mode3:      key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "playlists")),
	Mode4:      key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "channels")),
	listKeyMap: list.DefaultKeyMap(),
	searchInputMap: textinput.KeyMap{
		CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b")),
		WordForward:             key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f")),
		WordBackward:            key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b")),
		DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w")),
		DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d")),
		DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k")),
		DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u")),
		DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d")),
		LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a")),
		LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),
		AcceptSuggestion:        key.NewBinding(key.WithKeys("right", "ctrl+tab"), key.WithHelp("→/right", "autocomplete suggestion")),
		NextSuggestion:          key.NewBinding(key.WithKeys("down", "ctrl+n")),
		PrevSuggestion:          key.NewBinding(key.WithKeys("up", "ctrl+p")),
	},
}

// TODO show all key bindings from subcomponents (list, textinput)

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.listKeyMap.CloseFullHelp, k.listKeyMap.Filter}, // k.Refresh
		{k.Play, k.Switch, k.listKeyMap.ClearFilter},              // k.searchInputMap.AcceptSuggestion, k.Delete
		//{k.Mode1, k.Mode2, k.Mode3, k.Mode4},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.listKeyMap.Filter, k.Switch, k.listKeyMap.ShowFullHelp} // k.Refresh
}
