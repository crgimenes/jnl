# Publish Command - Hugo Export

The `publish` command allows you to export journal entries that have the `public` tag to Hugo format with proper frontmatter.

## Basic Usage

```bash
# Export to specific directory
jnl publish /path/to/hugo/content/posts

# Use configuration from Lua file
jnl publish

# Use --path flag
jnl publish --path /path/to/hugo/content/posts
```

## Configuration in jnl_init.lua

```lua
-- Publication settings
PublishPath = "~/Documents/hugo-blog/content/posts"  -- Target directory
PublishTag = "public"                                 -- Tag that marks publishable entries
BlockedTags = "secret,private,confidential"          -- Tags that prevent publication
```

## Security Tags

The system has built-in protections against accidental publication of sensitive content:

### Publication Tag

- Default: `@public`
- Only entries with this tag will be considered for publication

### Blocked Tags

- Default: `secret`, `private`
- Entries with any of these tags will **never** be published, even if they have `@public`
- Useful for preventing leakage of confidential information

## Example of Publishable Entry

### Journal Entry

```markdown
;;; jnl
date: 2025-07-05T10:56:44-03:00
title: How to Optimize Websites
tags: @development, @public, @web
user: username
;;;

# How to Optimize Websites

This article explains optimization techniques...

## Main Techniques

1. CSS/JS Minification
2. Image Compression
3. CDN
```

### Generated Hugo Output

```markdown
+++
date = "2025-07-05T10:56:44-03:00"
lastmod = "2025-07-05T10:56:44-03:00"
title = "How to Optimize Websites"
description = "This article explains optimization techniques..."
tags = ["development", "web"]
+++

# How to Optimize Websites

This article explains optimization techniques...

## Main Techniques

1. CSS/JS Minification
2. Image Compression
3. CDN
```

## Example of Blocked Entry

```markdown
;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Confidential Meeting
tags: @company, @public, @secret
;;;

We discussed confidential strategies...
```

This entry **will not be published** because it contains the `@secret` tag, even though it has `@public`.

## Command Output

```
Publishing entries from /path/to/diary to /path/to/hugo
Looking for entries with tag 'public'...
✅ Published: 2025-07-05T10-56-49-optimization-tips.md
⚠️  Skipped 2025-07-05T10-58-15-secret-meeting.md: contains blocked tags

Publish summary:
  Published: 1 entries
  Skipped: 8 entries
```

## Customization

### Change Publication Tag

```lua
PublishTag = "blog"  -- Now uses @blog instead of @public
```

### Add More Blocked Tags

```lua
BlockedTags = "secret,private,confidential,company,client"
```

### Use Different Directories

```bash
jnl publish ~/blog/posts           # Personal blog
jnl publish ~/work-blog/content    # Professional blog
```

## Generated Hugo Frontmatter

The command automatically generates:

- **date**: Original entry date
- **lastmod**: Same date (can be customized later)
- **title**: From header or first line of content
- **description**: First sentence or first 150 characters
- **tags**: Clean tags (without @ and without the publication tag)

## Security

The system is designed to be secure by default:

1. **Opt-in**: Only explicitly marked entries are published
2. **Blacklist**: Blocked tags prevent accidental publication
3. **Visibility**: Clear report of what was/wasn't published
4. **Reversible**: Original files are never modified
