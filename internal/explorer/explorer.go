//

// Package explorer handles all filesystem interactions: reading directories,
// sorting entries, and launching external editors/apps.
package explorer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EntryType distinguishes files from directories.
type EntryType int

const (
	EntryTypeFile EntryType = iota
	EntryTypeDir
)

// Entry represents a single file or directory item in the explorer.
type Entry struct {
	Name    string
	Path    string
	Type    EntryType
	Size    int64
	ModTime time.Time
	IsHidden bool
}

// IsDir returns true if this entry is a directory.
func (e Entry) IsDir() bool {
	return e.Type == EntryTypeDir
}

// SizeString returns a human-readable file size.
func (e Entry) SizeString() string {
	if e.IsDir() {
		return "—"
	}
	switch {
	case e.Size < 1024:
		return fmt.Sprintf("%dB", e.Size)
	case e.Size < 1024*1024:
		return fmt.Sprintf("%.1fK", float64(e.Size)/1024)
	case e.Size < 1024*1024*1024:
		return fmt.Sprintf("%.1fM", float64(e.Size)/(1024*1024))
	default:
		return fmt.Sprintf("%.1fG", float64(e.Size)/(1024*1024*1024))
	}
}

// ReadDir reads a directory and returns sorted entries.
// Directories come first, then files, both sorted alphabetically.
// When showHidden is false, entries whose names begin with "." are omitted.
func ReadDir(dirPath string, showHidden bool) ([]Entry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory %q: %w", dirPath, err)
	}

	var result []Entry
	for _, de := range entries {
		isHidden := strings.HasPrefix(de.Name(), ".")
		if isHidden && !showHidden {
			continue // skip dotfiles/dotdirs when hidden view is off
		}

		info, err := de.Info()
		if err != nil {
			continue // skip entries we can't stat
		}

		entry := Entry{
			Name:     de.Name(),
			Path:     filepath.Join(dirPath, de.Name()),
			ModTime:  info.ModTime(),
			IsHidden: isHidden,
		}

		if de.IsDir() {
			entry.Type = EntryTypeDir
		} else {
			entry.Type = EntryTypeFile
			entry.Size = info.Size()
		}

		result = append(result, entry)
	}

	// Sort: directories first, then files; each group alphabetically (case-insensitive)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type == EntryTypeDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

// OpenAction represents an action to perform on an entry.
type OpenAction int

const (
	ActionOpen   OpenAction = iota // open directory or VS Code for file
	ActionVSCode                   // open with VS Code
	ActionCursor                   // open with Cursor
	ActionNvim                     // open with Neovim
	ActionVim                      // open with Vim
	ActionNano                     // open with Nano
)

// Perform executes the given action on the entry.
// Terminal-based editors run in the current process (replacing the TUI temporarily).
// GUI apps are launched detached.
func Perform(entry Entry, action OpenAction) error {
	switch action {
	case ActionOpen:
		if entry.IsDir() {
			// Handled by the UI layer (navigate into dir) — callers should check.
			return nil
		}
		return launchDetached("code", entry.Path)

	case ActionVSCode:
		return launchDetached("code", entry.Path)

	case ActionCursor:
		return launchDetached("cursor", entry.Path)

	case ActionNvim:
		return launchTerminal("nvim", entry.Path)

	case ActionVim:
		return launchTerminal("vim", entry.Path)

	case ActionNano:
		return launchTerminal("nano", entry.Path)
	}
	return nil
}

// launchDetached opens a GUI application without blocking.
func launchDetached(app, path string) error {
	cmd := exec.Command(app, path)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	return cmd.Start()
}

// launchTerminal opens a terminal editor, blocking until it exits.
// It connects stdin/stdout/stderr so the editor has full terminal control.
func launchTerminal(app, path string) error {
	cmd := exec.Command(app, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RenameEntry renames a file or directory.
func RenameEntry(oldPath, newName string) (string, error) {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("rename failed: %w", err)
	}
	return newPath, nil
}

// CopyEntry copies a file or directory recursively.
// For files, it copies content. For directories, it does a recursive copy.
func CopyEntry(srcPath, dstDir string) error {
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	baseName := filepath.Base(srcPath)
	dstPath := filepath.Join(dstDir, copyName(baseName, dstDir))

	if info.IsDir() {
		return copyDir(srcPath, dstPath)
	}
	return copyFile(srcPath, dstPath)
}

// copyName generates a non-colliding name in dstDir by appending " copy", " copy 2", etc.
func copyName(name, dir string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	candidate := base + " copy" + ext
	if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
		return candidate
	}
	for i := 2; ; i++ {
		candidate = fmt.Sprintf("%s copy %d%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, _ := os.Stat(src)
	return os.WriteFile(dst, data, info.Mode())
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}