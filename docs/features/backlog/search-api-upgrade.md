# Feature: Search API upgrade (fresh content for news and current events)

**Status: Backlog**

## Problem

The current `web_search` tool uses the **DuckDuckGo Instant Answer API** (no API key). That API returns instant answers and Wikipedia-style results only; for many news or current-events queries it returns **empty**. The agent then falls back to `web_get` on news homepages, which often return JS-rendered HTML and truncated or unusable content. The tools effectively “battle” to get fresh content and lose.

## Proposal

### 1. Use a real search API (with optional API key)

Replace or supplement the DuckDuckGo Instant Answer backend with a search provider that returns **real web/search results** (snippets + URLs), and optionally **news-focused** results:

| Option | API key | Notes |
|--------|--------|--------|
| **Google Custom Search** | Yes (`GOOGLE_API_KEY` + `GOOGLE_CSE_ID`) | Full web + optional news; well-documented; free tier limited queries/day. |
| **Serper** (serper.dev) | Yes (`SERPER_API_KEY`) | Google results via API; single key; good for news; paid tiers. |
| **Tavily** | Yes (`TAVILY_API_KEY`) | Built for AI agents; search + optional “news” focus; free tier available. |
| **Bing Web Search** | Yes (`BING_SEARCH_API_KEY`) | Web + news; Azure portal. |

Recommendation: support **one primary provider** (e.g. Serper or Tavily for simplicity), with **DuckDuckGo Instant Answer as fallback** when no key is set, so the product works out of the box but improves when the user supplies a key.

### 2. Configuration

- **Env**: e.g. `AI_ASSISTANT_SEARCH_PROVIDER=duckduckgo|serper|tavily|google` and provider-specific keys (`SERPER_API_KEY`, `TAVILY_API_KEY`, `GOOGLE_API_KEY` + `GOOGLE_CSE_ID`, etc.).
- **Config**: Optional `SearchProvider` and API key fields on server config, loaded from env.
- No API keys in code or repo; keys only from env (or future secrets mechanism).

### 3. Optional: dedicated news tool

- A separate tool **`news_search`** that calls a news-specific API (e.g. NewsAPI.org, or the news endpoint of the same provider) can improve “what’s happening in X” queries.
- Alternatively, a single `web_search` that accepts an optional `news_only` (or `type=news`) parameter and uses the provider’s news endpoint when available.

### 4. Optional: improve `web_get` for article content

- When the agent fetches a URL, many news sites return HTML that is hard to use (JS-rendered or noisy).
- Optional: run fetched HTML through an **HTML-to-text** or **readability** step (e.g. go-readability or similar) and return cleaned article text so that when the agent follows a link from search, it gets readable content instead of raw HTML.

## Summary

- **Yes, using specific API keys for Google (or Serper/Tavily/Bing) search is the main lever** to get fresh content: real search results with snippets and links, and optionally news-focused results.
- Keep DuckDuckGo Instant Answer as default when no key is set; when a key is configured, use the chosen provider so the assistant can answer current-events questions reliably.

## Acceptance criteria (draft)

- [ ] Config/env for search provider and provider-specific API keys (no keys in code).
- [ ] At least one “full” search provider implemented (e.g. Serper or Tavily) returning web results (title, snippet, URL) for arbitrary queries.
- [ ] When no key is set, `web_search` continues to use DuckDuckGo Instant Answer (current behaviour).
- [ ] When a key is set, `web_search` uses the configured provider; tool description/documentation updated so the LLM knows search can return news/current-events results when configured.
