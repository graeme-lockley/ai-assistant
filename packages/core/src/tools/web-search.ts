import { Type } from "@mariozechner/pi-ai";
import type { AgentTool } from "@mariozechner/pi-agent-core";

const TavilyURL = "https://api.tavily.com/search";

export function webSearchTool(tavilyApiKey: string): AgentTool {
  return {
    name: "web_search",
    label: "Web search",
    description: "Search the web via Tavily. Requires TAVILY_API_KEY on the server.",
    parameters: Type.Object({
      query: Type.String({ description: "Search query" }),
    }),
    execute: async (_id, params) => {
      const p = params as { query: string };
      if (!tavilyApiKey) {
        return {
          content: [{ type: "text", text: "error: TAVILY_API_KEY is not set" }],
          details: {},
        };
      }
      const body = {
        query: p.query,
        search_depth: "basic",
        max_results: 10,
      };
      const res = await fetch(TavilyURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${tavilyApiKey}`,
        },
        body: JSON.stringify(body),
      });
      const text = await res.text();
      if (!res.ok) {
        return {
          content: [{ type: "text", text: `web_search error: ${res.status} ${text}` }],
          details: { status: res.status },
        };
      }
      return {
        content: [{ type: "text", text }],
        details: {},
      };
    },
  };
}
