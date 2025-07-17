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
- ⚙️ **Lua Configuration**: Flexible configuration with preprocessing and postprocessing hooks

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

JNL uses Lua for configuration, providing powerful customization options. Create a `jnl_init.lua` file in your current directory or at `~/.config/jnl/init.lua`.

### Basic Configuration

```lua
-- Journal settings
JournalPath = "~/Documents/journal"
TagPrefix = "@"

-- Publishing settings
PublishPath = "~/blog/content/posts"
PublishTag = "public"
BlockedTags = "secret,private,confidential"
```

### Directory-Based Automatic Tags

Configure automatic tags based on your workspace location:

```lua
DirectoryTags = {
    ["~/Documents/work"] = {"work", "secret"},
    ["~/Documents/personal"] = {"personal"},
    ["~/Documents/projects/client-a"] = {"client", "confidential"},
    ["~/Documents/blog"] = {"blog", "public"},
}
```

**How it works:**

- When you create a journal entry in `~/Documents/work/project1/`, it automatically adds `@work` and `@secret` tags
- Subdirectories inherit their parent's tags
- Perfect for maintaining security boundaries and organization

### Custom Hooks

```lua
-- Preprocessing hook - modify content before editing
function PreProc(text)
    return text
end

-- Postprocessing hook - modify content after editing
function PostProc(text)
    return text
end

-- Post-save hook - called after the journal entry has been saved
-- Receives the full path to the saved file and its content
function PostSave(filePath, content)
    -- Example: check for specific tags
    if string.find(content, "@public") then
        print("Entry marked as public - ready for publishing!")
        -- Copy to public directory
        -- os.execute("cp '" .. filePath .. "' ~/blog/posts/")
    end
    
    -- Example: backup important entries
    if string.find(content, "@important") then
        os.execute("cp '" .. filePath .. "' '" .. filePath .. ".important.bak'")
    end
    
    -- Example: word count statistics
    local wordCount = 0
    for word in content:gmatch("%S+") do
        wordCount = wordCount + 1
    end
    print("Word count: " .. wordCount)
end

-- Custom editor execution
function Exec(editor, file)
    local cmd = string.format('%s %s', editor, file)
    return os.execute(cmd)
end
```

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

```lua
DirectoryTags = {
    ["~/Documents/work"] = {"work", "secret"},
    ["~/Documents/clients"] = {"client", "confidential"},
}
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
