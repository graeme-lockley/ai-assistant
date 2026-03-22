import { Type } from "@mariozechner/pi-ai";
import type { AgentTool } from "@mariozechner/pi-agent-core";

const MAX_BYTES = 512 * 1024;

export function webGetTool(): AgentTool {
  return {
    name: "web_get",
    label: "Fetch URL",
    description: "HTTP GET a URL and return response body as text (truncated).",
    parameters: Type.Object({
      url: Type.String({ description: "URL to fetch" }),
    }),
    execute: async (_id, params) => {
      const p = params as { url: string };
      const res = await fetch(p.url, {
        redirect: "follow",
        headers: { "User-Agent": "ai-assistant/0.1" },
      });
      const buf = await res.arrayBuffer();
      const slice = buf.byteLength > MAX_BYTES ? buf.slice(0, MAX_BYTES) : buf;
      let text: string;
      try {
        text = new TextDecoder("utf8", { fatal: false }).decode(slice);
      } catch {
        text = "[binary or undecodable content]";
      }
      if (buf.byteLength > MAX_BYTES) {
        text += "\n\n[truncated]";
      }
      return {
        content: [{ type: "text", text: `status ${res.status}\n\n${text}` }],
        details: { status: res.status },
      };
    },
  };
}
