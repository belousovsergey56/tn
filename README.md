# tn

Terminal note - A fast CLI note manager featuring search, Git, and Obsidian

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

## Contents
- [Motivation](#motivation)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Demo](#demo)
- [Dependencies](#dependencies)
- [Configuration file](#configuration-file)
- [Storage structure](#storage-structure)
- [Example template](#example-template)
- [Ideas](#ideas)

## Motivation
This is the second iteration of the utility (the first one was written in Python). I decided to rewrite it in Go to learn the language, boost performance, and expand its feature set.

As someone who loves working in the terminal and uses Obsidian, I appreciate that notes are stored as plain files in a specific directory. This allows me to read, edit, or add notes seamlessly from both the GUI app and the console. My goal was to eliminate context switching to a heavy UI, allowing me to instantly find information in my knowledge base or jot down a quick thought on the fly. 

Instead of a complex TUI (Terminal User Interface), I wanted a tool that leverages familiar utilities with a unified interface, integrating smoothly with `fzf`, `fd`, `rg` / `grep`, and editors like `nvim` or `hx`.

## Features
- Find and open notes in a terminal-based text editor
- Create a new note and open it immediately in your editor
- Support for note templates (just specify the template path in the config file)
- Interactive search and deletion of notes
- Inline notes: save quick thoughts directly to a timestamped file via command arguments
- Find and open notes directly in the Obsidian desktop app
- Automatic Git synchronization
- Manual Git synchronization

## Installation
Choose one of the following methods to install `tn`:

### 1. Using `go install` (For Go developers)
If you have Go installed on your system, you can build and install the binary directly from the source:
```bash
go install github.com/belousovsergey56/tn@latest
```
*Make sure your $GOPATH/bin (usually ~/go/bin) is added to your system's PATH.*

### 2. Quick Install Script (Via curl)
You can install the latest pre-compiled binary with a single command:
```bash
curl -sSfL curl -sSfL https://raw.githubusercontent.com/belousovsergey56/tn/main/install.sh | sh
```

### 3. Manual Download (GitHub Releases)
- Go to the [Releases](https://github.com/belousovsergey56/tn/releases) page.
- Download the archive matching your operating system and CPU architecture (e.g., tn_Linux_x86_64.tar.gz).
- Extract the archive and move the tn binary to a directory in your PATH (such as /usr/local/bin or ~/.local/bin).

## Usage
- `tn help` / `tn -h` / `tn --help` - Help about any command
- `tn config` or `tn c` - Open config file for edit
- `tn delete` or `tn d` - Interactive search and delete a note
- `tn edit` or `tn e` - Interactive search and edit a note in the terminal
- `tn grep` or `tn g` - Interactive full-text search across notes
- `tn [text]` or `tn inline [text]` or `tn i [text]` - Create a quick timestamped note directly from arguments
- `tn new [filename]` or `tn n [filename]` - Create a new note and open it in the terminal
- `tn open` or `tn o` - Interactive search and open a note in Obsidian
- `tn sync` or `tn s` - Manually synchronize your notes vault with the remote Git repository

## Demo
- Config
- Inline note
- Edit note
- Delete note
- Full-text search
- New note
- Open note in Obsidian
- Sync local and remote repository

## Dependencies
- github.com/BurntSushi/toml v1.6.0
- github.com/ktr0731/go-fuzzyfinder v0.9.0
- github.com/spf13/cobra v1.10.2

## Configuration file
```toml
# Storage mode: currently "files", planning to add a "db" mode for database storage later.
storage_mode = "files"

# Path to the directory where notes will be stored.
# Supports ~ or $HOME.
path_to_main_vault = "$HOME/terminal-note"

# Path to the directory where inline notes will be stored. 
# Supports ~ or $HOME.
path_to_inline_note = "$HOME/terminal-note/cli_note"

# File extension for text notes: e.g., .txt, .md 
file_extension = ".md"

# Path to the template file.
# Supports ~ or $HOME.
path_to_template_note = "$HOME/terminal-note/Templates/base.md"

# Editor to write notes comfortably: vi, vim, nvim, micro, nano, etc.
editor = "hx"
```

## Storage structure
```bash
➜ ~/my-note
.
├── cli_note
│   ├── 2026-06-01 17_44_34.md
│   ├── 2026-06-01 17_45_34.md
│   └── 2026-06-02 10_41_32.md
├── file name with space.md
├── networking.md
├── newgifile.md
├── one more file 4.md
├── one more file.md
├── research new work.md
└── Templates
    └── base.md
```

## Example template
```md
---
Creation date:
Modification date:
links:
tags:
---
```

## Ideas
- [ ] Import/Export from and to databases
- [ ] Implement database storage mode
