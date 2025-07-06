# Directory-Based Automatic Tags

This feature allows you to configure automatic tags that will be added to journal entries based on the directory where the journal is being created.

## Configuration

Add the `DirectoryTags` table to your `jnl_init.lua` file:

```lua
-- Directory-specific automatic tags
-- Maps directory paths to tags that should be automatically added
-- If the journal is created in that directory or any subdirectory, the tags will be added
DirectoryTags = {
    ["~/Documents/work"] = {"work", "secret"},           -- Work directory gets "work" and "secret" tags
    ["~/Documents/personal"] = {"personal"},             -- Personal directory gets "personal" tag
    ["~/Documents/projects/client-project"] = {"client", "secret"}, -- Client work gets special tags
    ["~/Documents/public"] = {"public", "blog"},         -- Public content directory
}
```

## How It Works

1. **Directory Matching**: When you create a new journal entry with `jnl add`, the system checks if the current working directory matches any of the configured directories.

2. **Inheritance**: The system also checks parent directories. If you're in a subdirectory of a configured directory, the tags will still be applied.

3. **Automatic Addition**: The configured tags are automatically added to the journal entry's header, along with the existing automatic tags (directory path components, git branch, etc.).

## Examples

### Example 1: Work Directory with Secret Tag

Configuration:

```lua
DirectoryTags = {
    ["~/Documents/work"] = {"work", "secret"}
}
```

When you run `jnl add` from `~/Documents/work/project1/`, the following tags will be automatically added:

- `@Documents` (from path)
- `@work` (from path)  
- `@project1` (from path)
- `@main` (git branch, if applicable)
- `@work` (from DirectoryTags - duplicate will be handled)
- `@secret` (from DirectoryTags)

### Example 2: Client Project with Restricted Publishing

Configuration:

```lua
DirectoryTags = {
    ["~/Documents/clients/acme-corp"] = {"client", "acme", "secret"}
}

BlockedTags = "secret,private,confidential"
```

Any journal entry created in `~/Documents/clients/acme-corp/` or its subdirectories will automatically receive the `@secret` tag, preventing it from being published even if you add `@public` manually.

### Example 3: Public Blog Directory

Configuration:

```lua
DirectoryTags = {
    ["~/Documents/blog"] = {"blog", "public"}
}
```

Journal entries created in the blog directory will automatically be tagged for publishing.

## Security Benefits

This feature is particularly useful for preventing accidental publication of sensitive content:

1. **Automatic Secret Tagging**: Set up work directories to automatically add `@secret` tags
2. **Client Separation**: Different client directories can have different confidentiality tags
3. **Default Privacy**: Make work-related directories private by default

## Path Resolution

- **Home Directory Expansion**: Paths starting with `~/` are automatically expanded
- **Absolute Paths**: The system converts relative paths to absolute paths for matching
- **Subdirectory Matching**: Works with any subdirectory depth

## Integration with Existing Features

This feature works seamlessly with:

- **Publish System**: Blocked tags prevent publishing
- **Existing Auto-Tags**: Directory-based tags are added to path and git branch tags
- **Manual Tags**: You can still add additional tags manually when creating entries

## Use Cases

1. **Work-Life Separation**: Automatically tag work-related entries as private
2. **Client Confidentiality**: Ensure client work is never accidentally published
3. **Project Organization**: Auto-tag entries by project or category
4. **Publishing Workflow**: Pre-configure directories for automatic publishing
5. **Legal Compliance**: Ensure sensitive directories always get appropriate classification tags
