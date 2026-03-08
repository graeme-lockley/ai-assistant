# Feature: Tool collection

**Status: Done**

## Summary

Add a fixed set of tools the assistant can call during a turn. The server sends tool definitions to the LLM; when the LLM returns tool calls, the server runs the requested tools and sends results back to the LLM, then returns the final reply to the client.

## Tools

| Tool | Purpose | Parameters |
|------|---------|-------------|
| **web_search** | Run a web search via Tavily and return snippets/links. Requires `TAVILY_API_KEY`. | `query` (string). |
| **web_get** | Fetch a URL and return the response body as text. | `url` (string). |
| **exec_bash** | Run a bash command. | `command` (string). |
| **read_file** | Read a file's contents. | `path` (string). |
| **read_dir** | List directory entries (names, optionally types). | `path` (string). |
| **write_file** | Create or overwrite a file. | `path` (string), `content` (string). |
| **merge_file** | Insert or replace a region in a file (by line range or markers). | `path`, `content`, `strategy`. |

## Paths and working directory

- **File tools** (**read_file**, **read_dir**, **write_file**, **merge_file**): Paths are resolved relative to a configurable **workspace root**. Default: `~/.ai-assistant.workspace`. Override via `AI_ASSISTANT_WORKSPACE` or `AI_ASSISTANT_ROOT_DIR`. Resolve to an absolute path and reject if the result lies outside the root (e.g. block `../` escape).
- **exec_bash**: Run with the workspace root as the **current working directory**. No allowlist or sandbox; keep the feature minimal.

## Integration

- **Protocol**: Extend requests/responses so the server can send tool definitions to the LLM and receive tool-call payloads; the server runs tools and sends tool results back in the same turn (or next request), then returns the final assistant message.
- **Server**: The component that handles a connection gets a **root directory** from config and a **tool runner** that implements the seven tools. When the LLM returns tool calls, it invokes the tool runner and feeds results back to the LLM.
- **Config**: Workspace root for file operations and for exec_bash cwd; **default** is `~/.ai-assistant.workspace`. Override via `AI_ASSISTANT_WORKSPACE` or `AI_ASSISTANT_ROOT_DIR`. Web search requires `TAVILY_API_KEY`.

## Out of scope

- Plugins or user-defined tools.
- Browser automation or display.
- Allowlists, sandboxes, or URL filtering.

## Acceptance criteria

- [x] All seven tools are implemented and callable by the server.
- [x] File tools take a path; resolution is relative to the configured workspace root (default: `~/.ai-assistant.workspace`); path traversal outside the root is rejected.
- [x] exec_bash runs with the configured root as cwd.
- [x] Protocol supports tool definitions, tool calls, and tool results; server runs tools and returns the final reply to the client.

## Implementation (done)

- **Config**: `AI_ASSISTANT_WORKSPACE` (preferred) or `AI_ASSISTANT_ROOT_DIR` for workspace root; default `~/.ai-assistant.workspace`. `TAVILY_API_KEY` required for web_search.
- **Tools package** (`internal/tools`): Runner interface; path resolution against workspace root with traversal rejection; web_search (Tavily Search API, POST with Bearer token; requires `TAVILY_API_KEY`), web_get, exec_bash, read_file, read_dir, write_file, merge_file (strategies: replace by line range, markers).
- **LLM** (`internal/llm`): Message supports ToolCalls and ToolCallID; ToolDefinitions() for OpenAI-format tools; CompleteStreamWithTools streams and returns tool calls; messagesToOpenAI builds API messages (empty content filled for tool/assistant-with-calls for API compatibility).
- **Agent** (`internal/agent`): Optional tool runner; when set, uses CompleteStreamWithTools and loops: run tools, append assistant + tool messages, re-call until final reply; tool call/result logging with correlation ID (tool call id).
- **Session/Server**: Store and server create tool runner from config root; pass runner into agent.
