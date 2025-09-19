package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	playingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	pausedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	stoppedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Background(lipgloss.Color("#44475A")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)
)

// Model represents the TUI model
type Model struct {
	player      *Player
	width       int
	height      int
	cursor      int
	showHelp    bool
	searchMode  bool
	searchQuery string
}

// NewModel creates a new TUI model
func NewModel(player *Player) *Model {
	return &Model{
		player: player,
		cursor: 0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchMode {
		return m.handleSearchInput(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "h":
		m.showHelp = !m.showHelp
		return m, nil

	case "/":
		m.searchMode = true
		m.searchQuery = ""
		return m, nil

	case " ":
		if m.player.State == Playing {
			m.player.Pause()
		} else if m.player.State == Paused {
			m.player.Resume()
		} else {
			m.player.Play()
		}
		return m, nil

	case "n":
		m.player.NextTrack()
		return m, nil

	case "p":
		m.player.PreviousTrack()
		return m, nil

	case "r":
		m.player.ToggleRepeat()
		return m, nil

	case "s":
		m.player.ToggleShuffle()
		return m, nil

	case "j", "down":
		if m.cursor < len(m.player.FilteredTracks)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "g":
		m.cursor = 0
		return m, nil

	case "G":
		m.cursor = len(m.player.FilteredTracks) - 1
		return m, nil

	case "enter":
		if len(m.player.FilteredTracks) > 0 && m.cursor < len(m.player.FilteredTracks) {
			m.player.CurrentTrack = m.cursor
			m.player.Play()
		}
		return m, nil

	}

	return m, nil
}

// handleSearchInput handles search mode input
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.player.Search(m.searchQuery)
		m.searchMode = false
		m.cursor = 0
		return m, nil

	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		return m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil

	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
		return m, nil
	}
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	if m.searchMode {
		return m.renderSearch()
	}

	return m.renderMain()
}

// renderMain renders the main interface
func (m Model) renderMain() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🎵 TuiTunes"))
	content.WriteString("\n\n")

	// Status bar
	content.WriteString(m.renderStatusBar())
	content.WriteString("\n\n")

	// Playlist
	content.WriteString(m.renderPlaylist())
	content.WriteString("\n\n")

	// Controls
	content.WriteString(m.renderControls())

	return content.String()
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	var status strings.Builder

	// what's happening
	var stateText string
	switch m.player.State {
	case Playing:
		stateText = playingStyle.Render("▶ PLAYING")
	case Paused:
		stateText = pausedStyle.Render("⏸ PAUSED")
	case Stopped:
		stateText = stoppedStyle.Render("⏹ STOPPED")
	}

	status.WriteString(stateText)
	status.WriteString(" ")

	// Current track info
	if track := m.player.GetCurrentTrack(); track != nil {
		status.WriteString(fmt.Sprintf("| %s - %s", track.Artist, track.Title))
	}

	// progress bar
	if m.player.State == Playing || m.player.State == Paused {
		status.WriteString("\n")
		status.WriteString(m.renderProgressBar())
	}

	// Repeat/Shuffle indicators
	if m.player.Repeat {
		status.WriteString(" | 🔁")
	}
	if m.player.Shuffle {
		status.WriteString(" | 🔀")
	}

	return statusBarStyle.Render(status.String())
}

// renderProgressBar renders the progress bar
func (m Model) renderProgressBar() string {
	if m.player.Streamer == nil {
		return ""
	}

	position := m.player.GetPosition()
	length := m.player.GetLength()

	if length == 0 {
		return ""
	}

	progress := float64(position) / float64(length)
	barWidth := 30
	filledWidth := int(progress * float64(barWidth))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	timeStr := fmt.Sprintf("%s / %s",
		formatDuration(position),
		formatDuration(length))

	return fmt.Sprintf("%s %s", bar, timeStr)
}

// renderPlaylist renders the track list
func (m Model) renderPlaylist() string {
	if len(m.player.FilteredTracks) == 0 {
		return "No tracks found. Press '/' to search or add music files to the directory."
	}

	var content strings.Builder
	content.WriteString("Playlist:\n")

	for i, track := range m.player.FilteredTracks {
		var style lipgloss.Style
		switch {
		case i == m.cursor:
			style = selectedStyle
		case i == m.player.CurrentTrack:
			style = playingStyle
		default:
			style = normalStyle
		}

		// Track info
		info := fmt.Sprintf("%s - %s", track.Artist, track.Title)
		if track.Album != "" {
			info += fmt.Sprintf(" (%s)", track.Album)
		}

		// Duration
		duration := formatDuration(track.Duration)

		// show what's playing
		var prefix string
		if i == m.player.CurrentTrack {
			switch m.player.State {
			case Playing:
				prefix = "▶ "
			case Paused:
				prefix = "⏸ "
			default:
				prefix = "⏹ "
			}
		} else {
			prefix = "  "
		}

		line := fmt.Sprintf("%s%s - %s", prefix, info, duration)
		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}

	return content.String()
}

// renderControls renders the control help
func (m Model) renderControls() string {
	controls := []string{
		"Space: Play/Pause",
		"N: Next",
		"P: Previous",
		"R: Repeat",
		"S: Shuffle",
		"/: Search",
		"H: Help",
		"Q: Quit",
	}

	return helpStyle.Render(strings.Join(controls, " | "))
}

// renderSearch renders the search interface
func (m Model) renderSearch() string {
	var content strings.Builder
	content.WriteString("Search: ")
	content.WriteString(m.searchQuery)
	content.WriteString("_")
	content.WriteString("\n\n")
	content.WriteString("Press Enter to search, Esc to cancel")
	return content.String()
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	help := `
TuiTunes - Terminal Music Player

CONTROLS:
  Space       Play/Pause
  N           Next track
  P           Previous track
  R           Toggle repeat mode
  S           Toggle shuffle
  /           Search music
  H           Show/hide this help
  Q           Quit

NAVIGATION:
  ↑/↓ or J/K  Navigate playlist
  G           Go to top of playlist
  G (Shift+G) Go to bottom of playlist
  Enter       Play selected track


SEARCH:
  /           Enter search mode
  Type        Enter search query
  Enter       Execute search
  Esc         Cancel search

SUPPORTED FORMATS:
  MP3, WAV, FLAC, M4A, AAC, OGG

Press H to return to the main interface.
`
	return helpStyle.Render(help)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
