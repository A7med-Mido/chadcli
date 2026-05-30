// internal/ui/styles.go

// Package ui/styles defines the complete visual theme using lipgloss.
package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — dark terminal theme.
var (
	colorBg         = lipgloss.Color("#1e1e2e") // base background
	colorSurface    = lipgloss.Color("#313244") // panel/menu surfaces
	colorOverlay    = lipgloss.Color("#45475a") // dividers, borders
	colorMuted      = lipgloss.Color("#6c7086") // subtle text
	colorText       = lipgloss.Color("#cdd6f4") // primary text
	colorSubtext    = lipgloss.Color("#a6adc8") // secondary text
	colorAccent     = lipgloss.Color("#89b4fa") // blue accent (directories)
	colorGreen      = lipgloss.Color("#a6e3a1") // files / success
	colorYellow     = lipgloss.Color("#f9e2af") // warnings / rename
	colorRed        = lipgloss.Color("#f38ba8") // errors
	colorPink       = lipgloss.Color("#f5c2e7") // copy indicator
	colorTeal       = lipgloss.Color("#94e2d5") // search highlight
	colorSelected   = lipgloss.Color("#585b70") // selection background
	colorCursorBg   = lipgloss.Color("#89b4fa") // cursor row background
	colorCursorFg   = lipgloss.Color("#1e1e2e") // cursor row text
)

// Styles is the central style registry for the TUI.
type Styles struct {
	// Layout chrome
	AppBorder   lipgloss.Style
	StatusBar   lipgloss.Style
	BreadCrumb  lipgloss.Style
	HelpBar     lipgloss.Style

	// List items
	ItemNormal   lipgloss.Style
	ItemSelected lipgloss.Style // non-cursor selected (e.g. copied item)
	ItemCursor   lipgloss.Style // cursor row
	ItemHidden   lipgloss.Style // hidden files

	// Type-specific decorations
	DirName  lipgloss.Style
	FileName lipgloss.Style
	DirIcon  lipgloss.Style
	FileIcon lipgloss.Style

	// Search
	SearchPrompt    lipgloss.Style
	SearchHighlight lipgloss.Style
	SearchNormal    lipgloss.Style

	// Menu
	MenuBorder   lipgloss.Style
	MenuTitle    lipgloss.Style
	MenuItem     lipgloss.Style
	MenuSelected lipgloss.Style

	// Status/info
	InfoSize    lipgloss.Style
	InfoModTime lipgloss.Style
	StatusOK    lipgloss.Style
	StatusError lipgloss.Style
	StatusCopy  lipgloss.Style

	// Rename input
	InputPrompt lipgloss.Style
	InputStyle  lipgloss.Style
}

// DefaultStyles returns the dark Catppuccin-inspired theme.
func DefaultStyles() Styles {
	s := Styles{}

	s.AppBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorOverlay).
		Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		Background(colorSurface).
		Foreground(colorSubtext).
		Padding(0, 1)

	s.BreadCrumb = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	s.HelpBar = lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true)

	s.ItemNormal = lipgloss.NewStyle().
		Foreground(colorText).
		Padding(0, 1)

	s.ItemSelected = lipgloss.NewStyle().
		Foreground(colorPink).
		Padding(0, 1)

	s.ItemCursor = lipgloss.NewStyle().
		Background(colorCursorBg).
		Foreground(colorCursorFg).
		Bold(true).
		Padding(0, 1)

	s.ItemHidden = lipgloss.NewStyle().
		Foreground(colorMuted)

	s.DirName = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	s.FileName = lipgloss.NewStyle().
		Foreground(colorText)

	s.DirIcon = lipgloss.NewStyle().
		Foreground(colorYellow)

	s.FileIcon = lipgloss.NewStyle().
		Foreground(colorGreen)

	s.SearchPrompt = lipgloss.NewStyle().
		Foreground(colorTeal).
		Bold(true)

	s.SearchHighlight = lipgloss.NewStyle().
		Foreground(colorTeal).
		Underline(true).
		Bold(true)

	s.SearchNormal = lipgloss.NewStyle().
		Foreground(colorText)

	s.MenuBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Background(colorSurface)

	s.MenuTitle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		MarginBottom(1)

	s.MenuItem = lipgloss.NewStyle().
		Foreground(colorSubtext).
		Padding(0, 1)

	s.MenuSelected = lipgloss.NewStyle().
		Background(colorSelected).
		Foreground(colorText).
		Bold(true).
		Padding(0, 1)

	s.InfoSize = lipgloss.NewStyle().
		Foreground(colorMuted)

	s.InfoModTime = lipgloss.NewStyle().
		Foreground(colorMuted)

	s.StatusOK = lipgloss.NewStyle().
		Foreground(colorGreen)

	s.StatusError = lipgloss.NewStyle().
		Foreground(colorRed)

	s.StatusCopy = lipgloss.NewStyle().
		Foreground(colorPink)

	s.InputPrompt = lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true)

	s.InputStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Background(colorSurface).
		Padding(0, 1)

	return s
}