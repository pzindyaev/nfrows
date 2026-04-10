package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding // gg (handled as sequence in app.go)
	GotoBottom   key.Binding
	PrevTab      key.Binding
	NextTab      key.Binding
	Enter        key.Binding
	Back         key.Binding
	Add          key.Binding
	Delete       key.Binding
	Edit         key.Binding
	Flush        key.Binding
	Refresh      key.Binding
	Tab          key.Binding
	ShiftTab     key.Binding
	Quit         key.Binding
	Confirm      key.Binding
	Cancel       key.Binding
	Help         key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	GotoTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("gg", "go to top"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to bottom"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("h/←", "prev tab"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("l/→", "next tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Add: key.NewBinding(
		key.WithKeys("a", "o"),
		key.WithHelp("a/o", "add"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e", "i"),
		key.WithHelp("e/i", "edit"),
	),
	Flush: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "flush"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n/esc", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
