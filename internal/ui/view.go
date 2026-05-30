package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/A7med-Mido/chadcli/internal/icons"
	"github.com/A7med-Mido/chadcli/internal/search"
)

// View renders the complete TUI frame.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteByte('\n')
	b.WriteString(m.renderDivider())
	b.WriteByte('\n')
	b.WriteString(m.renderList())
	b.WriteByte('\n')
	b.WriteString(m.renderDivider())
	b.WriteByte('\n')
	b.WriteString(m.renderBottom())
	b.WriteByte('\n')
	b.WriteString(m.renderHelp())

	content := b.String()

	if m.state == stateMenu {
		return m.renderMenuOverlay(content)
	}
	return content
}

// ─── Header ──────────────────────────────────────────────────────────────────

func (m *Model) renderHeader() string {
	path := m.currentDir
	if home, err := os.UserHomeDir(); err == nil {
		if rel, err2 := filepath.Rel(home, path); err2 == nil && !strings.HasPrefix(rel, "..") {
			path = "~/" + rel
		}
	}

	crumb := m.styles.BreadCrumb.Render(" " + path)

	total := len(m.entries)
	shown := len(m.filtered)
	var countStr string
	if m.lastQuery != "" {
		countStr = fmt.Sprintf(" %d/%d matches ", shown, total)
	} else if m.showHidden {
		countStr = fmt.Sprintf(" %d items (+ hidden) ", total)
	} else {
		countStr = fmt.Sprintf(" %d items ", total)
	}
	count := m.styles.InfoSize.Render(countStr)

	var clip string
	if m.clipboard != nil {
		clip = m.styles.StatusCopy.Render(fmt.Sprintf("   %s", m.clipboard.Name))
	}

	available := m.width - lipgloss.Width(crumb) - lipgloss.Width(count) - lipgloss.Width(clip) - 2
	padding := ""
	if available > 0 {
		padding = strings.Repeat(" ", available)
	}
	return crumb + padding + clip + count
}

// ─── Divider ─────────────────────────────────────────────────────────────────

func (m *Model) renderDivider() string {
	style := lipgloss.NewStyle().Foreground(colorOverlay)
	w := m.width
	if w <= 0 {
		w = 80
	}
	return style.Render(strings.Repeat("─", w))
}

// ─── File list ───────────────────────────────────────────────────────────────

func (m *Model) renderList() string {
	listH := m.listHeight()
	if len(m.filtered) == 0 {
		empty := lipgloss.NewStyle().Foreground(colorMuted).Italic(true).
			Width(m.width).Align(lipgloss.Center).
			Height(listH).
			Render("No matches")
		return empty
	}

	start, end := scrollWindow(m.cursor, len(m.filtered), listH)

	var lines []string
	for i := start; i < end; i++ {
		lines = append(lines, m.renderItem(i, m.filtered[i]))
	}
	for len(lines) < listH {
		lines = append(lines, strings.Repeat(" ", m.width))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderItem(idx int, result search.Result) string {
	entry := result.Entry
	isCursor := idx == m.cursor

	ico := icons.Icon(entry.Name, entry.IsDir())
	var iconStr string
	if entry.IsDir() {
		iconStr = m.styles.DirIcon.Render(ico)
	} else {
		iconStr = m.styles.FileIcon.Render(ico)
	}

	var typeBadge string
	if entry.IsDir() {
		typeBadge = lipgloss.NewStyle().Foreground(colorAccent).Faint(true).Render("📁  dir")
	} else {
		typeBadge = lipgloss.NewStyle().Foreground(colorGreen).Faint(true).Render("📄 file")
	}

	var nameStr string
	if m.lastQuery != "" && len(result.Positions) > 0 {
		nameStr = highlightName(entry.Name, result.Positions, m.styles, entry.IsDir())
	} else if entry.IsDir() {
		nameStr = m.styles.DirName.Render(entry.Name)
	} else if entry.IsHidden {
		nameStr = m.styles.ItemHidden.Render(entry.Name)
	} else {
		nameStr = m.styles.FileName.Render(entry.Name)
	}

	sizeStr := m.styles.InfoSize.Render(fmt.Sprintf("%8s", entry.SizeString()))
	modStr := m.styles.InfoModTime.Render(entry.ModTime.Format("Jan 02 15:04"))

	left := fmt.Sprintf("  %s %s  %s", typeBadge, iconStr, nameStr)
	right := fmt.Sprintf("%s  %s  ", sizeStr, modStr)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := m.width - leftW - rightW
	if gap < 1 {
		gap = 1
	}
	row := left + strings.Repeat(" ", gap) + right

	if isCursor {
		rowW := lipgloss.Width(row)
		if rowW < m.width {
			row += strings.Repeat(" ", m.width-rowW)
		}
		return m.styles.ItemCursor.Render(row)
	}
	return row
}

func highlightName(name string, positions []int, s Styles, isDir bool) string {
	posSet := make(map[int]bool, len(positions))
	for _, p := range positions {
		posSet[p] = true
	}
	var b strings.Builder
	for i, r := range []rune(name) {
		ch := string(r)
		if posSet[i] {
			b.WriteString(s.SearchHighlight.Render(ch))
		} else if isDir {
			b.WriteString(s.DirName.Render(ch))
		} else {
			b.WriteString(s.SearchNormal.Render(ch))
		}
	}
	return b.String()
}

// ─── Bottom bar ───────────────────────────────────────────────────────────────

func (m *Model) renderBottom() string {
	switch m.state {
	case stateSearch:
		prompt := m.styles.SearchPrompt.Render("  Search: ")
		return prompt + m.searchInput.View()
	case stateRename:
		prompt := m.styles.InputPrompt.Render(fmt.Sprintf("  Rename %q → ", m.renameEntry.Name))
		return prompt + m.renameInput.View()
	default:
		if m.statusMsg != "" {
			if m.statusIsError {
				return m.styles.StatusError.Render("  ✖ " + m.statusMsg)
			}
			return m.styles.StatusOK.Render("  ✔ " + m.statusMsg)
		}
		return m.styles.HelpBar.Render("  Ready")
	}
}

// ─── Help bar ────────────────────────────────────────────────────────────────

func (m *Model) renderHelp() string {
	var parts []string
	switch m.state {
	case stateSearch:
		parts = []string{keyHint("↑↓", "navigate"), keyHint("enter", "confirm"), keyHint("esc", "cancel")}
	case stateMenu:
		parts = []string{keyHint("↑↓", "select"), keyHint("1-6", "shortcut"), keyHint("esc", "close")}
	case stateRename:
		parts = []string{keyHint("enter", "confirm"), keyHint("esc", "cancel")}
	default:
		parts = []string{
			keyHint("↑↓", "navigate"), keyHint("enter", "menu"), keyHint("/", "search"),
			keyHint("r", "rename"), keyHint("c/v", "copy/paste"), keyHint("esc/←", "go back"),
			keyHint(".", "toggle hidden"), keyHint("~", "home"), keyHint("q", "quit"),
		}
	}
	return m.styles.HelpBar.Render("  " + strings.Join(parts, "   "))
}

func keyHint(key, desc string) string {
	k := lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(key)
	d := lipgloss.NewStyle().Foreground(colorMuted).Render(" " + desc)
	return k + d
}

// ─── Menu overlay ────────────────────────────────────────────────────────────

func (m *Model) renderMenuOverlay(_ string) string {
	entry := m.menuEntry
	title := m.styles.MenuTitle.Render(
		fmt.Sprintf("%s  %s", icons.Icon(entry.Name, entry.IsDir()), entry.Name),
	)
	var rows []string
	for i, item := range m.menuItems {
		label := fmt.Sprintf("%d  %s  %s", i+1, item.icon, item.label)
		if i == m.menuCursor {
			rows = append(rows, m.styles.MenuSelected.Render(label))
		} else {
			rows = append(rows, m.styles.MenuItem.Render(label))
		}
	}
	menu := m.styles.MenuBorder.Render(title + "\n" + strings.Join(rows, "\n"))

	menuW := lipgloss.Width(menu)
	menuH := lipgloss.Height(menu)
	x := (m.width - menuW) / 2
	y := (m.height - menuH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
		lipgloss.NewStyle().MarginLeft(x).MarginTop(y).Render(menu),
		lipgloss.WithWhitespaceChars(" "),
	)
}

// ─── Utilities ───────────────────────────────────────────────────────────────

func scrollWindow(cursor, total, height int) (int, int) {
	if total <= height {
		return 0, total
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
	}
	return start, end
}