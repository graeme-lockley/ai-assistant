# Feature: Control plane (slash commands)

**Status: Done**

## Summary

Introduce a control plane reachable from the REPL via slash commands. Users can exit the session, list and select models, and view help. Some commands are handled entirely by the client; others instruct the server.

## Slash commands

| Command | Handler | Description |
|---------|---------|-------------|
| **/exit** | Server | Leave the interaction and terminate the session. The REPL sends a session-close signal to the server, then exits. |
| **/models** | Server | Show the models that are available. The client sends a request that instructs the server to return the list of available models; the REPL displays them. |
| **/model** [*name*] | Server | Select a model for this session. The client sends a request to set the session's model; the server associates the chosen model with the session for subsequent turns. If *name* is omitted, show the current model for the session. |
| **/help** | Client | Show the slash commands and brief usage. Handled entirely in the REPL; no server round-trip. |

## Behaviour

### /exit

- **REPL**: When the user types `/exit`, the REPL treats it as a control command (not a normal message).
- **Client**: Sends a request to the server to close the session (e.g. existing mechanism: `X-Session-Close: true` with `X-Session-Id`; response 204 No Content).
- **REPL**: After the close request completes (or fails), the REPL terminates and the user leaves the interaction.

### /models

- **REPL**: When the user types `/models`, the client sends a request that instructs the server to return the list of available models.
- **Server**: Exposes a way to answer "what models are available?" (e.g. dedicated endpoint, or a special request type on the same API). Returns the list (e.g. model IDs and optional display names). **For now the list of models is hardcoded**; support for multiple models (e.g. from config or API) will be added later.
- **REPL**: Displays the list to the user in a readable form.

### /model

- **REPL**: When the user types `/model` or `/model <name>`, the client sends a request to set or query the session model.
- **Server**: Associates the chosen model with the session (keyed by session ID). Subsequent turns in that session use this model. If the request omits a model name, the server returns the current model for the session.
- **REPL**: Shows confirmation (e.g. "Model set to …" or "Current model: …").

### /help

- **REPL**: When the user types `/help`, the REPL prints the list of slash commands and short usage text. No request is sent to the server.

## Integration

- **Protocol**: Define how the client requests "list models" and "set/query session model" (e.g. new endpoints, or request types/headers on the existing API). Reuse existing session identification (`X-Session-Id`) and session close (`X-Session-Close: true`) where applicable.
- **Server**: The component that handles sessions stores per-session model selection and uses it when calling the LLM. The list of available models is **hardcoded for now**; multiple models (from config or LLM API) will be added later.
- **REPL**: Parses input for leading `/exit`, `/models`, `/model`, `/help` and dispatches to client-side handling (/help) or server request (/exit, /models, /model).

## Out of scope

- Slash commands for tool configuration or other control-plane features in this iteration.
- Authentication or authorization for model selection.
- Per-request model override (session-level only for v1).
- Dynamic model list from API or config (v1 uses a hardcoded list; multiple models later).

## Acceptance criteria

- [x] **/exit**: User can type `/exit` in the REPL; client sends session close to the server; REPL then exits.
- [x] **/models**: User can type `/models`; client requests available models from the server; REPL displays the list.
- [x] **/model**: User can type `/model <name>` to set the session model; server stores it for the session and uses it for subsequent turns. User can type `/model` with no argument to see the current model.
- [x] **/help**: User can type `/help`; REPL prints slash commands and usage without contacting the server.
