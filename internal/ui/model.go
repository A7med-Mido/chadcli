// internal/ui/model.go

// Package ui contains the Bubble Tea TUI model, view, and update logic.
// Package ui contains the Bubble Tea TUI model, view, and update logic.
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/A7med-Mido/chadcli/internal/explorer"
	"github.com/A7med-Mido/chadcli/internal/search"
)

// ─── Application states ──────────────────────────────────────────────────────

type appState int

const (
	stateExplorer appState = iota
	stateSearch
	stateMenu
	stateRename
)

// ─── Messages ────────────────────────────────────────────────────────────────

type dirLoadedMsg struct {
	entries []explorer.Entry
	err     error
}

type statusFlashMsg struct {
	text    string
	isError bool
}

type editorFinishedMsg struct{ err error }

// ─── Model ───────────────────────────────────────────────────────────────────

// Model is the root Bubble Tea model.
type Model struct {
	styles Styles

	// Filesystem state
	currentDir string
	entries    []explorer.Entry
	history    []string

	// Filtered view
	filtered []search.Result
	cursor   int

	// State machine
	state appState

	// Search
	searchInput textinput.Model
	lastQuery   string

	// Action menu
	menuItems  []menuItem
	menuCursor int
	menuEntry  explorer.Entry

	// Rename
	renameInput textinput.Model
	renameEntry explorer.Entry

	// Copy/Paste clipboard
	clipboard *explorer.Entry

	// Dotfile visibility (toggled with ".")
	showHidden bool

	// Status bar
	statusMsg     string
	statusIsError bool

	// Terminal dimensions
	width  int
	height int
}

type menuItem struct {
	label  string
	action explorer.OpenAction
	icon   string
}

var defaultMenuItems = []menuItem{
	{label: "Open", action: explorer.ActionOpen, icon: ""},
	{label: "VS Code", action: explorer.ActionVSCode, icon: ""},
	{label: "Cursor", action: explorer.ActionCursor, icon: ""},
	{label: "Neovim", action: explorer.ActionNvim, icon: ""},
	{label: "Vim", action: explorer.ActionVim, icon: ""},
	{label: "Nano", action: explorer.ActionNano, icon: ""},
}

// NewModel constructs the initial model for startDir.
func NewModel(startDir string) (*Model, error) {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	si := textinput.New()
	si.Placeholder = "fuzzy search…"
	si.CharLimit = 120
	si.Width = 40

	ri := textinput.New()
	ri.CharLimit = 255
	ri.Width = 40

	m := &Model{
		styles:      DefaultStyles(),
		currentDir:  absDir,
		searchInput: si,
		renameInput: ri,
		menuItems:   defaultMenuItems,
		width:       120,
		height:      40,
	}

	entries, err := explorer.ReadDir(absDir, false)
	if err != nil {
		return nil, err
	}
	m.entries = entries
	m.rebuildFiltered()
	return m, nil
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m *Model) Init() tea.Cmd {
	return nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case statusFlashMsg:
		m.statusMsg = msg.text
		m.statusIsError = msg.isError
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.setStatus("editor error: "+msg.err.Error(), true)
		}
		return m, m.loadDir(m.currentDir)

	case dirLoadedMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
			return m, nil
		}
		m.entries = msg.entries
		m.rebuildFiltered()
		m.clampCursor()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateExplorer:
		return m.handleExplorerKey(msg)
	case stateSearch:
		return m.handleSearchKey(msg)
	case stateMenu:
		return m.handleMenuKey(msg)
	case stateRename:
		return m.handleRenameKey(msg)
	}
	return m, nil
}

// ── Explorer ──────────────────────────────────────────────────────────────────

func (m *Model) handleExplorerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		// If a search filter is active, clear it first; otherwise go up a directory.
		if m.lastQuery != "" {
			m.searchInput.SetValue("")
			m.lastQuery = ""
			m.rebuildFiltered()
			m.cursor = 0
			return m, nil
		}
		return m, m.navigateUp()
	case tea.KeyUp:
		m.moveCursor(-1)
	case tea.KeyDown:
		m.moveCursor(1)
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = max(0, len(m.filtered)-1)
	case tea.KeyPgUp:
		m.moveCursor(-(m.listHeight() / 2))
	case tea.KeyPgDown:
		m.moveCursor(m.listHeight() / 2)
	case tea.KeyBackspace:
		return m, m.navigateUp()
	case tea.KeyEnter:
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.menuEntry = m.filtered[m.cursor].Entry
		m.menuCursor = 0
		m.state = stateMenu
		return m, nil
	case tea.KeyRunes:
		switch msg.String() {
		case "q", "Q":
			return m, tea.Quit
		case "/":
			m.state = stateSearch
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			return m, textinput.Blink
		case ".":
			// Toggle dotfile visibility and reload the current directory.
			m.showHidden = !m.showHidden
			if m.showHidden {
				m.setStatus("Showing hidden files", false)
			} else {
				m.setStatus("Hiding dotfiles", false)
			}
			return m, m.loadDir(m.currentDir)
		case "c":
			if len(m.filtered) > 0 {
				e := m.filtered[m.cursor].Entry
				cp := e
				m.clipboard = &cp
				m.setStatus(fmt.Sprintf("Copied  %s", e.Name), false)
			}
		case "v":
			if m.clipboard == nil {
				m.setStatus("Nothing to paste — press c to copy first", true)
				return m, nil
			}
			return m, m.doPaste()
		case "r":
			if len(m.filtered) > 0 {
				m.renameEntry = m.filtered[m.cursor].Entry
				m.renameInput.SetValue(m.renameEntry.Name)
				m.renameInput.Focus()
				m.state = stateRename
				return m, textinput.Blink
			}
		case "l":
			if len(m.filtered) > 0 {
				e := m.filtered[m.cursor].Entry
				if e.IsDir() {
					return m, m.navigateInto(e.Path)
				}
			}
		case "~":
			if home, err := os.UserHomeDir(); err == nil {
				return m, m.navigateInto(home)
			}
		}
	}
	return m, nil
}

// ── Search ────────────────────────────────────────────────────────────────────

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateExplorer
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.lastQuery = ""
		m.rebuildFiltered()
		m.cursor = 0
		return m, nil
	case tea.KeyEnter:
		m.state = stateExplorer
		m.searchInput.Blur()
		return m, nil
	case tea.KeyUp:
		m.moveCursor(-1)
		return m, nil
	case tea.KeyDown:
		m.moveCursor(1)
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		q := m.searchInput.Value()
		if q != m.lastQuery {
			m.lastQuery = q
			m.rebuildFiltered()
			m.cursor = 0
		}
		return m, cmd
	}
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func (m *Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateExplorer
	case tea.KeyUp:
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case tea.KeyDown:
		if m.menuCursor < len(m.menuItems)-1 {
			m.menuCursor++
		}
	case tea.KeyEnter:
		return m.executeMenuAction()
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if unicode.IsDigit(r) {
				idx := int(r-'0') - 1
				if idx >= 0 && idx < len(m.menuItems) {
					m.menuCursor = idx
					return m.executeMenuAction()
				}
			}
		}
	}
	return m, nil
}

// ── Rename ────────────────────────────────────────────────────────────────────

func (m *Model) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateExplorer
		m.renameInput.Blur()
		return m, nil
	case tea.KeyEnter:
		newName := strings.TrimSpace(m.renameInput.Value())
		m.state = stateExplorer
		m.renameInput.Blur()
		if newName == "" || newName == m.renameEntry.Name {
			return m, nil
		}
		if _, err := explorer.RenameEntry(m.renameEntry.Path, newName); err != nil {
			m.setStatus("Rename failed: "+err.Error(), true)
		} else {
			m.setStatus(fmt.Sprintf("Renamed → %s", newName), false)
		}
		return m, m.loadDir(m.currentDir)
	default:
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd
	}
}

// ─── Commands ────────────────────────────────────────────────────────────────

func (m *Model) loadDir(path string) tea.Cmd {
	showHidden := m.showHidden
	return func() tea.Msg {
		entries, err := explorer.ReadDir(path, showHidden)
		return dirLoadedMsg{entries: entries, err: err}
	}
}

func (m *Model) navigateInto(path string) tea.Cmd {
	m.history = append(m.history, m.currentDir)
	m.currentDir = path
	m.cursor = 0
	m.searchInput.SetValue("")
	m.lastQuery = ""
	m.state = stateExplorer
	return m.loadDir(path)
}

func (m *Model) navigateUp() tea.Cmd {
	if len(m.history) > 0 {
		prev := m.history[len(m.history)-1]
		m.history = m.history[:len(m.history)-1]
		m.currentDir = prev
	} else {
		parent := filepath.Dir(m.currentDir)
		if parent == m.currentDir {
			return nil
		}
		m.currentDir = parent
	}
	m.cursor = 0
	m.searchInput.SetValue("")
	m.lastQuery = ""
	m.state = stateExplorer
	return m.loadDir(m.currentDir)
}

func (m *Model) executeMenuAction() (tea.Model, tea.Cmd) {
	m.state = stateExplorer
	item := m.menuItems[m.menuCursor]
	entry := m.menuEntry

	// Navigate into directory
	if item.action == explorer.ActionOpen && entry.IsDir() {
		return m, m.navigateInto(entry.Path)
	}

	// Terminal editors — suspend TUI, run editor, reload
	if item.action == explorer.ActionNvim ||
		item.action == explorer.ActionVim ||
		item.action == explorer.ActionNano {
		cmd := buildEditorCmd(item.action, entry.Path)
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})
	}

	// GUI / detached applications
	if err := explorer.Perform(entry, item.action); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err), true)
	} else {
		m.setStatus(fmt.Sprintf("Opened with %s", item.label), false)
	}
	return m, nil
}

func (m *Model) doPaste() tea.Cmd {
	clip := m.clipboard
	dir := m.currentDir
	return func() tea.Msg {
		if err := explorer.CopyEntry(clip.Path, dir); err != nil {
			return statusFlashMsg{text: "Paste failed: " + err.Error(), isError: true}
		}
		return statusFlashMsg{text: "Pasted: " + clip.Name, isError: false}
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (m *Model) rebuildFiltered() {
	m.filtered = search.Filter(m.lastQuery, m.entries)
}

func (m *Model) clampCursor() {
	if len(m.filtered) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) moveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor += delta
	m.clampCursor()
}

func (m *Model) setStatus(msg string, isError bool) {
	m.statusMsg = msg
	m.statusIsError = isError
}

func (m *Model) listHeight() int {
	reserved := 9
	h := m.height - reserved
	if h < 5 {
		return 5
	}
	return h
}

func buildEditorCmd(action explorer.OpenAction, path string) *exec.Cmd {
	switch action {
	case explorer.ActionNvim:
		return exec.Command("nvim", path)
	case explorer.ActionVim:
		return exec.Command("vim", path)
	case explorer.ActionNano:
		return exec.Command("nano", path)
	}
	return exec.Command("vim", path)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}