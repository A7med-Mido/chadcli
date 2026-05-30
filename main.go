// main.go

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// entry represents a file or directory in the current folder
type entry struct {
	name    string
	isDir   bool
	absPath string
}

// model holds the application state
type model struct {
	entries      []entry         // all entries in current directory
	filtered     []entry         // entries matching the filter
	cursor       int             // current selection index in filtered list
	filter       string          // search filter text
	width        int             // terminal width
	height       int             // terminal height
	status       string          // status message (errors, reloads)
	needsRefresh bool            // flag to reload directory entries
}

// initialModel reads the current directory and builds the initial state
func initialModel() (model, error) {
	entries, err := readCurrentDir()
	if err != nil {
		return model{}, fmt.Errorf("failed to read current directory: %w", err)
	}

	m := model{
		entries:  entries,
		filtered: entries,
		cursor:   0,
		filter:   "",
		status:   "Ready — ↑/↓ navigate, Enter open in VS Code, type to filter, r refresh, Esc clear, q quit",
	}
	m.applyFilter()
	return m, nil
}

// readCurrentDir reads the current working directory and returns sorted entries
func readCurrentDir() ([]entry, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(cwd)
	if err != nil {
		return nil, err
	}

	entries := make([]entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		absPath := filepath.Join(cwd, de.Name())
		entries = append(entries, entry{
			name:    de.Name(),
			isDir:   de.IsDir(),
			absPath: absPath,
		})
	}

	// Sort: directories first, then files; both groups alphabetically
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir // directories come first
		}
		return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
	})

	return entries, nil
}

// applyFilter rebuilds filtered entries based on the current filter string (case-insensitive)
func (m *model) applyFilter() {
	if m.filter == "" {
		m.filtered = m.entries
	} else {
		lowerFilter := strings.ToLower(m.filter)
		filtered := make([]entry, 0, len(m.entries))
		for _, e := range m.entries {
			if strings.Contains(strings.ToLower(e.name), lowerFilter) {
				filtered = append(filtered, e)
			}
		}
		m.filtered = filtered
	}

	// Adjust cursor if needed
	if len(m.filtered) == 0 {
		m.cursor = -1
	} else if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	} else if m.cursor < 0 && len(m.filtered) > 0 {
		m.cursor = 0
	}
}

// refresh reloads the directory contents and reapplies the filter
func (m *model) refresh() tea.Cmd {
	newEntries, err := readCurrentDir()
	if err != nil {
		m.status = fmt.Sprintf("Error refreshing: %v", err)
		return nil
	}
	m.entries = newEntries
	m.applyFilter()
	m.status = "Reloaded directory"
	return nil
}

// openSelected opens the currently selected file/directory in VS Code
func (m *model) openSelected() tea.Cmd {
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		m.status = "Nothing selected"
		return nil
	}
	selected := m.filtered[m.cursor]
	cmd := exec.Command("code", selected.absPath)
	err := cmd.Run()
	if err != nil {
		m.status = fmt.Sprintf("Failed to open '%s': %v", selected.name, err)
	} else {
		m.status = fmt.Sprintf("Opened '%s' in VS Code", selected.name)
	}
	return nil
}

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles keyboard events and other messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Clear status on any user action (except errors that we set again)
		m.status = ""

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyRunes:
			// Quit on 'q' or Ctrl+C
			if msg.String() == "q" || msg.String() == "Q" {
				return m, tea.Quit
			}
			// Refresh on 'r'
			if msg.String() == "r" || msg.String() == "R" {
				m.refresh()
				return m, nil
			}
			// Regular character input for filtering
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.applyFilter()
				return m, nil
			}
			return m, nil

		case tea.KeyUp:
			if len(m.filtered) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.filtered) - 1 // wrap around
				}
			}
			return m, nil

		case tea.KeyDown:
			if len(m.filtered) > 0 {
				m.cursor++
				if m.cursor >= len(m.filtered) {
					m.cursor = 0 // wrap around
				}
			}
			return m, nil

		case tea.KeyEnter:
			m.openSelected()
			return m, nil

		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}
			return m, nil

		case tea.KeyDelete:
			// optional: same as backspace for convenience
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}
			return m, nil

		case tea.KeyEsc:
			if m.filter != "" {
				m.filter = ""
				m.applyFilter()
			}
			return m, nil

		default:
			return m, nil
		}
	}
	return m, nil
}

// View renders the UI
func (m model) View() string {
	if m.width == 0 {
		return "Loading...\n"
	}

	var b strings.Builder

	// Filter line
	fmt.Fprintf(&b, "🔍 Filter: %s\n", m.filter)
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", m.width))

	// Status line
	fmt.Fprintf(&b, "📌 %s\n", m.status)
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", m.width))

	// Help line
	help := "↑/↓: move • Enter: open in VS Code • Type: filter • r: refresh • Esc: clear • q: quit"
	fmt.Fprintf(&b, "💡 %s\n", help)
	fmt.Fprintf(&b, "💡 %s\n", "press ctrl+C to copy the file/dir name")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", m.width))

	// List area dimensions
	listHeight := m.height - 5 // subtract lines used for filter, status, help, separators
	if listHeight < 1 {
		listHeight = 1
	}

	if len(m.filtered) == 0 {
		fmt.Fprintf(&b, "  No matching files or directories\n")
		return b.String()
	}

	// Determine scroll window
	start, end := calculateScroll(m.cursor, listHeight, len(m.filtered))

	// Render visible entries
	for i := start; i < end; i++ {
		e := m.filtered[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "→ "
		}

		displayName := e.name
		if e.isDir {
			displayName += "/"
		}

		// Truncate long names if necessary (rough)
		maxNameLen := m.width - 4
		if len(displayName) > maxNameLen {
			displayName = displayName[:maxNameLen-3] + "..."
		}

		fmt.Fprintf(&b, "%s%s\n", prefix, displayName)
	}

	// If there are more items beyond scroll, show indicator
	if end < len(m.filtered) {
		fmt.Fprintf(&b, "\n  ... and %d more", len(m.filtered)-end)
	}

	return b.String()
}

// calculateScroll returns start and end indices for a scrollable list
func calculateScroll(cursor, height, total int) (int, int) {
	if total <= height {
		return 0, total
	}
	// Center cursor if possible
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	if start+height > total {
		start = total - height
	}
	return start, start + height
}

func main() {
	m, err := initialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}