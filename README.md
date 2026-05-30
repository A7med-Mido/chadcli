# ChadCli

A keyboard-driven terminal file explorer written in Go, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```
~/projects/chadcli                                                         32 items
──────────────────────────────────────────────────────────────────────────────────────
  dir    cmd          Jan 14 09:12
  dir    internal     Jan 14 09:12
  file   go.mod       3.1K   Jan 14 09:12
  file   go.sum       18.4K  Jan 14 09:12
  file  󰈙 README.md    4.2K   Jan 14 09:12
──────────────────────────────────────────────────────────────────────────────────────
  Ready
  ↑↓ navigate   enter menu   / search   r rename   c/v copy/paste   h go up   q quit
```

---

## Features

| Feature | Details |
|---|---|
| **Navigation** | Arrow keys, Page Up/Down, Home/End, vim-style `h`/`l` |
| **Entry menu** | Press `Enter` to open: Open, VS Code, Cursor, Neovim, Vim, Nano |
| **Fuzzy search** | Press `/`, type — real-time filtering with typo tolerance and matched character highlights |
| **Icons** | Nerd Font icons for 50+ file types and special directories |
| **Colors** | Catppuccin Mocha theme; directories in blue, files in green, hidden files dimmed |
| **Copy/Paste** | `c` to copy, `v` to paste (auto-names to avoid collisions) |
| **Rename** | `r` to rename inline |
| **Smart scrolling** | List scrolls, keeping cursor centred |

---

## Project Architecture

```
chadcli/
├── cmd/
│   └── main.go           # Entry point: parse args, create model, run Bubble Tea
├── internal/
│   ├── explorer/
│   │   └── explorer.go       # Filesystem: ReadDir, Entry, RenameEntry, CopyEntry, Perform
│   ├── icons/
│   │   └── icons.go          # Maps filenames/extensions → Nerd Font Unicode icons
│   ├── search/
│   │   └── search.go         # Fuzzy filtering via sahilm/fuzzy
│   └── ui/
│       ├── model.go          # Bubble Tea Model: state machine, Update, commands
│       ├── view.go           # Bubble Tea View: renders header, list, overlays
│       └── styles.go         # Lipgloss style definitions (colour palette + components)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Architecture decisions

**Bubble Tea (MVC)** — the `Model` struct in `model.go` owns all application state.
`Update` is a pure function that returns `(Model, Cmd)`. `View` in `view.go` is a pure
function that renders the model to a string. Side effects (directory reads, editor
launches) happen exclusively in `Cmd` functions that return `Msg` values back into
`Update`. This makes the explorer easy to reason about and test.

**State machine** — the model has an explicit `appState` field (`stateExplorer`,
`stateSearch`, `stateMenu`, `stateRename`). Each state has its own key handler, so
key bindings never collide and new states can be added without touching existing logic.

**Internal packages** — `explorer`, `icons`, and `search` have no dependencies on
the TUI layer, making them independently testable and reusable.

---

## Requirements

- Go 1.22+
- A terminal with [Nerd Font](https://www.nerdfonts.com/) support for icons
  (e.g. FiraCode Nerd Font, JetBrains Mono Nerd Font)

### Recommended terminals

| Terminal | Notes |
|---|---|
| [Kitty](https://sw.kovidgoyal.net/kitty/) | Full Nerd Font + 24-bit colour |
| [WezTerm](https://wezfurlong.org/wezterm/) | Full Nerd Font + ligatures |
| [iTerm2](https://iterm2.com/) | macOS; set a Nerd Font in preferences |
| [Windows Terminal](https://aka.ms/terminal) | Set a Nerd Font face |

---

## Shell integration — cd on quit

When you press `q`, the shell will `cd` to whatever directory you were browsing. This requires a one-time setup because a child process can't change the parent shell's directory directly — the wrapper handles that via a temp file.

### bash / zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/chadcli/shell/chadcli.sh
```

Then use `gx` instead of `chadcli`:

```bash
gx            # open in current directory
gx ~/projects # open in a specific directory
```

### fish

Copy the fish function file (fish auto-loads files whose name matches the function):

```fish
cp /path/to/chadcli/shell/chadcli.fish ~/.config/fish/functions/gx.fish
```

Then use `gx` the same way.

### How it works

1. The `gx` wrapper creates a temp file and passes its path via `$CHADCLI_CD_FILE`.
2. When the binary exits, `main.go` writes `model.CurrentDir()` into that file.
3. The wrapper reads the file and runs `cd` in the parent shell process.
4. The temp file is always cleaned up, even if the binary crashes.

---

### From source

```bash
git clone https://github.com/you/chadcli
cd chadcli
make install          # installs to $(go env GOPATH)/bin/goxplorer
```

Or build a local binary:

```bash
make build            # creates ./dist/goxplorer
./dist/chadcli      # run from current directory
./dist/chadcli ~/   # or pass a starting directory
```

---

## Keyboard Reference

### Explorer (normal mode)

| Key | Action |
|---|---|
| `↑` / `↓` | Move cursor |
| `Page Up` / `Page Down` | Move half a page |
| `Home` / `End` | Jump to first / last item |
| `Enter` | Open action menu |
| `Esc` / `Backspace` | Go up one directory (if no search active; clears search first) |
| `l` | Navigate into selected directory |
| `~` | Go to home directory |
| `/` | Start fuzzy search |
| `.` | Toggle dotfile / hidden file visibility |
| `r` | Rename selected item |
| `c` | Copy selected item to clipboard |
| `v` | Paste clipboard into current directory |
| `q` / `Ctrl+C` | Quit |

### Search mode (press `/`)

| Key | Action |
|---|---|
| Type anything | Filter in real time |
| `↑` / `↓` | Navigate filtered results |
| `Enter` | Confirm selection, return to explorer |
| `Esc` | Cancel search, restore full list |

### Action menu (press `Enter`)

| Key | Action |
|---|---|
| `↑` / `↓` | Move selection |
| `Enter` | Execute selected action |
| `1` – `6` | Shortcut to action by number |
| `Esc` | Close menu |

---

## Dependencies

| Library | Purpose |
|---|---|
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Elm-architecture TUI framework |
| [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) | Text input component |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Style / colour / layout |
| [sahilm/fuzzy](https://github.com/sahilm/fuzzy) | Fuzzy string matching |

---

## Extending chadcli

### Add a new editor

In `internal/explorer/explorer.go`, add a new `OpenAction` constant and handle it in
`Perform`. In `internal/ui/model.go`, add a `menuItem` to `defaultMenuItems`.

### Add file icons

In `internal/icons/icons.go`, add a new `case` to the `ext` switch in `fileIcon()`.
All icons use [Nerd Font](https://www.nerdfonts.com/cheat-sheet) code points.

### Change the colour theme

All colors and styles live in `internal/ui/styles.go`. Swap the `color*` variables
to adopt any terminal palette (e.g. Solarized, Gruvbox, Nord).