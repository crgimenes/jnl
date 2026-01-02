# JNL - Journal Tool

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue.svg)](https://golang.org/doc/install)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful command-line journal tool that helps you create, organize, and publish journal entries with intelligent automatic tagging and built-in security controls.

## Features

- 📝 **Smart Journal Creation**: Create entries with automatic metadata and tagging
- 🏷️ **Intelligent Tagging**: Automatic tags based on directory structure, git branch, and custom configurations
- 📂 **Directory-Based Tags**: Automatic categorization based on your workspace location
- 🔒 **Security Controls**: Prevent accidental publishing of sensitive content with blocked tags
- 📖 **Entry Management**: List, edit, view, and search your journal entries
- 🚀 **Hugo Publishing**: Export entries to Hugo static site generator format
- ⚙️ **Filo Configuration**: Flexible configuration with preprocessing and postprocessing hooks using the safe Filo scripting language

## Installation

### From Source

```bash
git clone https://github.com/crgimenes/jnl.git
cd jnl
go build -o jnl .
```

### Requirements

- Go 1.24 or higher
- Git (optional, for automatic branch tagging)

## Quick Start

1. **Create your first journal entry:**

   ```bash
   jnl add
   ```

2. **List your entries:**

   ```bash
   jnl ls
   ```

3. **View an entry:**

   ```bash
   jnl cat filename.md
   ```

## Commands

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `add` | Create a new journal entry | `jnl add --title "Meeting Notes"` |
| `ls` | List journal entries | `jnl ls "*.md"` |
| `edit` | Edit an existing entry | `jnl edit filename.md` |
| `cat` | Display entry content | `jnl cat filename.md` |
| `less` | View entry with header info | `jnl less filename.md` |
| `path` | Show journal directory path | `jnl path` |
| `version` | Show version information | `jnl version` |

### Publishing Commands

| Command | Description | Example |
|---------|-------------|---------|
| `publish` | Export entries to Hugo format | `jnl publish ~/blog/content` |

### Command Options

#### Add Command

- `--title, -t`: Set entry title
- `--force, -f`: Force save even if no changes detected
- `--commit, -c`: Run git commit after saving

#### List Command

- `--full-path, -f`: Show full file paths
- `pattern`: Filter entries by filename pattern

#### Publish Command

- `--path, -p`: Specify target directory for publishing

## Configuration

JNL uses the Filo scripting language for configuration, providing safe and powerful customization. Create a `jnl_init.filo` file in your current directory or at `~/.config/jnl/init.filo`.

### Basic Configuration

```lisp
;;; JNL Configuration File - Filo Format

(let ()
  ;; Journal settings
  (set JournalPath "~/Documents/journal")
  (set JournalTitlePrefix "")
  (set TagPrefix "@")
  (set ListenAddr ":8080")

  ;; Publishing settings
  (set PublishPath "~/blog/content/posts")
  (set PublishTag "public")
  (set BlockedTags "secret,private,confidential"))
```

### Directory-Based Automatic Tags

Configure automatic tags based on your workspace location:

```lisp
(set DirectoryTags
  (list
    (list "~/Documents/work" (list "work" "secret"))
    (list "~/Documents/personal" (list "personal"))
    (list "~/Documents/projects/client-a" (list "client" "confidential"))
    (list "~/Documents/blog" (list "blog" "public"))))
```

**How it works:**

- When you create a journal entry in `~/Documents/work/project1/`, it automatically adds `@work` and `@secret` tags
- Subdirectories inherit their parent's tags
- Perfect for maintaining security boundaries and organization

### Custom Hooks

```lisp
;; Preprocessing hook - modify content before editing
(def pre-proc (fn (text)
  text))

;; Postprocessing hook - modify content after editing  
(def post-proc (fn (text)
  text))

;; Post-save hook - called after journal entry is saved
;; Receives path and content, can trigger side effects
(def post-save (fn (file-path content)
  (if (str-find "@public" content)
      (let ()
        (jnl:exec (str-concat "cp '" file-path "' ~/blog/posts/"))
        (print "● Entry marked as public!"))
      #t)
  #t))

;; Custom editor execution (optional)
(def exec (fn (editor file)
  (jnl:exec (str-concat editor " " file))))
```

> **Note**: In Filo, `if` accepts 2 or 3 arguments: `(if cond then [else])`. For hooks like `post-save`, it’s usually best to always provide an `else` branch (often `#t`) so the function returns a boolean consistently.

JNL also registers a few helpful builtins for config scripts:

- `str-*` string functions (via Filo string builtins)
- `print` for debugging output
- `jnl:exec` to run an external command (expects exactly 1 string argument)

## Journal Entry Format

JNL uses a structured format with metadata headers:

```markdown
;;; jnl
date: 2025-07-06T15:30:00-03:00
dir: ~/Documents/work/project1
user: username
branch: main
tags: @work, @project1, @main, @secret
title: Meeting Notes
;;;

# Project Status Meeting

## Attendees
- Alice
- Bob

## Discussion Points
- Feature roadmap
- Budget planning
```

## Security Features

### Automatic Security Tagging

Configure directories to automatically add security tags:

```lisp
(set DirectoryTags
  (list
    (list "~/Documents/work" (list "work" "secret"))
    (list "~/Documents/clients" (list "client" "confidential"))))
```

### Publishing Controls

- **Publish Tag**: Only entries with the configured publish tag (default: `public`) are exported
- **Blocked Tags**: Entries with blocked tags (default: `secret`, `private`) are never published, even with a publish tag
- **Security First**: Prevents accidental exposure of sensitive information

Example:

```markdown
tags: @work, @public, @secret
```

This entry will **NOT** be published because it contains a blocked tag (`secret`), even though it has the publish tag (`public`).

## Hugo Integration

Export your journal entries to Hugo static site generator format:

```bash
jnl publish ~/blog/content/posts
```

**Features:**

- Automatic Hugo frontmatter generation
- Tag conversion and cleanup
- Date formatting
- Title extraction
- Security filtering

**Generated Hugo frontmatter:**

```toml
+++
date = "2025-07-06T15:30:00-03:00"
lastmod = "2025-07-06T15:30:00-03:00"
title = "Meeting Notes"
description = "Project status meeting with team members..."
tags = ["work", "project1", "meeting"]
+++
```

## Examples

### Daily Work Journal

```bash
cd ~/Documents/work
jnl add --title "Daily Standup"
# Automatically gets @work and @secret tags
```

### Personal Blog Entry

```bash
cd ~/Documents/blog
jnl add --title "My Trip to Japan"
# Automatically gets @blog and @public tags
```

### Client Work

```bash
cd ~/Documents/clients/acme-corp
jnl add --title "Project Requirements"
# Automatically gets @client and @confidential tags
```

## Use Cases

- **Software Development**: Track project progress, meeting notes, and technical decisions
- **Personal Journaling**: Daily reflections, travel logs, and life events
- **Work Documentation**: Client communications, project status, and team updates
- **Blog Writing**: Draft and publish blog posts with automatic Hugo integration
- **Research Notes**: Academic research, book notes, and study materials

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Created by [Cesar Gimenes](https://github.com/crgimenes)
