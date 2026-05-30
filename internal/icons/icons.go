// internal/icons/icons.go

// Package icons maps file names and extensions to Nerd Font / Unicode icons.
// Falls back to a generic icon when no specific match exists.
package icons




import (
	"path/filepath"
	"strings"
)

// Icon returns an icon string for the given filename.
// Pass isDir=true for directories.
func Icon(name string, isDir bool) string {
	if isDir {
		return dirIcon(name)
	}
	return fileIcon(name)
}

func dirIcon(name string) string {
	lower := strings.ToLower(name)
	switch lower {
	case ".git":
		return "󰊢"
	case "node_modules":
		return ""
	case ".github":
		return ""
	case "dist", "build", "out", "bin":
		return ""
	case "docs", "doc", "documentation":
		return "󰈙"
	case "src", "source":
		return ""
	case "test", "tests", "__tests__", "spec":
		return ""
	case "assets", "static", "public":
		return ""
	case "images", "img", "icons":
		return ""
	case "config", "configs", ".config":
		return ""
	case "scripts", "script":
		return ""
	case "vendor":
		return ""
	default:
		return ""
	}
}

// fileIcon returns an icon for a file based on its name or extension.
func fileIcon(name string) string {
	lower := strings.ToLower(name)

	// Exact name matches first
	switch lower {
	case "dockerfile", "dockerfile.dev", "dockerfile.prod":
		return ""
	case "docker-compose.yml", "docker-compose.yaml":
		return ""
	case "makefile", "gnumakefile":
		return ""
	case ".gitignore", ".gitattributes", ".gitmodules":
		return ""
	case ".env", ".env.local", ".env.example":
		return ""
	case "readme.md", "readme.txt", "readme":
		return ""
	case "license", "license.md", "license.txt":
		return ""
	case "package.json":
		return ""
	case "package-lock.json":
		return ""
	case "yarn.lock":
		return ""
	case "go.sum":
		return ""
	case "cargo.toml", "cargo.lock":
		return ""
	case "requirements.txt", "pipfile":
		return ""
	case "composer.json":
		return ""
	case "gemfile":
		return ""
	case ".bashrc", ".bash_profile", ".zshrc", ".zshenv", ".profile":
		return ""
	case "tsconfig.json":
		return ""
	case ".eslintrc", ".eslintrc.js", ".eslintrc.json":
		return "󰱺"
	case ".prettierrc", ".prettierrc.json":
		return ""
	case "jest.config.js", "jest.config.ts":
		return ""
	case "vite.config.js", "vite.config.ts":
		return ""
	case "webpack.config.js":
		return "󰜫"
	}

	// Extension-based matching
	ext := strings.TrimPrefix(filepath.Ext(lower), ".")
	switch ext {
	// Go
	case "go":
		return ""
	case "mod", "sum":
		return ""

	// Web / JS ecosystem
	case "js", "mjs", "cjs":
		return ""
	case "ts", "mts", "cts":
		return ""
	case "jsx":
		return ""
	case "tsx":
		return ""
	case "vue":
		return ""
	case "svelte":
		return ""
	case "html", "htm":
		return ""
	case "css":
		return ""
	case "scss", "sass":
		return ""
	case "less":
		return ""

	// Python
	case "py", "pyw", "pyx":
		return ""
	case "ipynb":
		return ""

	// Rust
	case "rs":
		return ""
	case "tom":
		return ""

	// Ruby
	case "rb", "rbw":
		return ""
	case "erb":
		return ""

	// Java / JVM
	case "java":
		return ""
	case "kt", "kts":
		return ""
	case "scala":
		return ""
	case "groovy":
		return ""
	case "class", "jar":
		return ""

	// C / C++
	case "c":
		return ""
	case "cpp", "cc", "cxx", "c++":
		return ""
	case "h", "hpp", "hxx":
		return ""

	// C#
	case "cs":
		return "󰌛"
	case "csproj", "sln":
		return "󰌛"

	// PHP
	case "php":
		return ""

	// Swift
	case "swift":
		return ""

	// Kotlin already covered above

	// Shell / Scripts
	case "sh", "bash":
		return ""
	case "zsh":
		return ""
	case "fish":
		return ""
	case "ps1", "psm1":
		return ""

	// Data / Config
	case "json":
		return ""
	case "yaml", "yml":
		return ""
	case "xml":
		return "󰗀"
	case "toml":
		return ""
	case "ini", "cfg", "conf":
		return ""
	case "env":
		return ""
	case "csv":
		return ""
	case "sql":
		return ""
	case "graphql", "gql":
		return ""
	case "proto":
		return ""

	// Docs / Text
	case "md", "mdx":
		return ""
	case "txt":
		return ""
	case "rst":
		return ""
	case "pdf":
		return ""
	case "doc", "docx":
		return ""
	case "xls", "xlsx":
		return ""
	case "ppt", "pptx":
		return ""

	// Images
	case "png", "jpg", "jpeg", "gif", "webp", "avif":
		return ""
	case "svg":
		return "󰜡"
	case "ico":
		return ""
	case "bmp", "tiff", "tif":
		return ""

	// Audio / Video
	case "mp3", "wav", "ogg", "flac", "aac", "m4a":
		return ""
	case "mp4", "avi", "mov", "mkv", "webm":
		return ""

	// Archives
	case "zip", "tar", "gz", "bz2", "xz", "7z", "rar":
		return ""

	// Binaries / Executables
	case "exe", "msi":
		return ""
	case "so", "dll", "dylib":
		return ""

	// Fonts
	case "ttf", "otf", "woff", "woff2":
		return ""

	// Lock files
	case "lock":
		return ""

	// Terraform / Infra
	case "tf", "tfvars":
		return ""

	// Dart / Flutter
	case "dart":
		return ""

	// Elixir
	case "ex", "exs":
		return ""

	// Haskell
	case "hs", "lhs":
		return ""

	// Lua
	case "lua":
		return ""

	// R
	case "r", "rmd":
		return ""

	// Clojure
	case "clj", "cljs", "cljc":
		return ""

	// Erlang
	case "erl", "hrl":
		return ""

	// Vim
	case "vim":
		return ""

	// Nix
	case "nix":
		return ""

	default:
		return ""
	}
}