# JNL - Journal Tool

A command-line journal tool that helps you create, organize, and publish journal entries with automatic tagging and publishing features.

## Features

- **Automatic Tagging**: Tags based on directory structure, git branch, and custom configurations
- **Publishing System**: Export entries to Hugo format with security controls
- **Lua Configuration**: Flexible configuration with preprocessing and postprocessing
- **Directory-Based Tags**: Automatic tags based on where you create entries
- **Security Controls**: Prevent accidental publishing of sensitive content

## Directory-Based Automatic Tags

You can configure automatic tags that will be added based on the directory where you create journal entries. This is useful for:

- **Security**: Automatically tag work-related entries as `secret` to prevent accidental publishing
- **Organization**: Auto-categorize entries by project or type
- **Workflow**: Pre-configure directories for automatic publishing or privacy

### Configuration

Add a `DirectoryTags` table to your `jnl_init.lua`:

```lua
DirectoryTags = {
    ["~/Documents/work"] = {"work", "secret"},
    ["~/Documents/personal"] = {"personal", "public"},
    ["~/Documents/projects/client-a"] = {"client", "secret"},
}
```

### How it Works

- The system checks if your current directory (or any parent directory) matches a configured path
- Matching tags are automatically added to the journal entry header
- Works alongside existing automatic tags (path components, git branch, etc.)

Example: Creating a journal in `~/Documents/work/project1/` automatically adds `@work` and `@secret` tags.
