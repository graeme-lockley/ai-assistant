# TOOLS

Built-in (code). Paths relative to workspace root. Loader: `context/routing/tools.json`. Optional `tools/<name>/TOOL.md`. Run `ai-assistant index` to regenerate cards.

## List

- **web_search** — Facts, definitions. Not news. `query`
- **web_get** — Fetch URL as text. News, articles. `url`
- **exec_bash** — Bash; cwd = workspace root. `command`
- **read_file** — Read file. `path`
- **read_dir** — List dir entries. `path`
- **write_file** — Create/overwrite; mkdir parents. `path`, `content`
- **merge_file** — Replace region: `replace` (start/end) or `markers` (begin/end_marker). `path`, `content`, `strategy` + line args

Workspace notes: this file or `tools/<name>/TOOL.md`.
