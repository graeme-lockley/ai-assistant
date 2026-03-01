# Feature: REPL history

**Status: Done**

## Summary

Add readline-style history to the REPL so users can recall and edit previous input using the arrow keys. Up/down navigate through history; left/right move within the current line as today.

## Behaviour

- **Up arrow**: Move to the previous entry in history (older).
- **Down arrow**: Move to the next entry in history (newer), or to the current line being edited if at the bottom.
- **Left/Right arrows**: Unchanged — move cursor within the current line.
- History is persisted across REPL sessions (e.g. to a file in the user config directory) so previous runs are available.
- Duplicate consecutive entries are not stored; the most recent of a run is kept.
- History length is bounded (e.g. last N entries or last N lines in the history file); configurable optional.

## Integration

- **REPL**: The readline (or equivalent) input layer loads and saves history; handles Up/Down for history navigation and Left/Right for line editing.
- **Config**: Optional path for history file; optional max history size. Sensible defaults (e.g. `~/.config/ai-assistant/repl_history` or similar, max 1000 entries).

## Out of scope

- Search within history (e.g. Ctrl+R).
- Per-project or per-workspace history (single global history is enough for v1).

## Acceptance criteria

- [x] Up/Down arrow keys navigate backwards and forwards through REPL input history.
- [x] Left/Right arrow keys continue to move the cursor within the current line.
- [x] History is persisted to a file and restored when the REPL starts.
- [x] Consecutive duplicate lines are not stored.
- [x] History size is bounded by a configurable limit (default documented).
