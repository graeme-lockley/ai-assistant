import type { AgentTool } from "@mariozechner/pi-agent-core";
import { execBashTool } from "./exec-bash.js";
import { fileOpTools } from "./file-ops.js";
import { webGetTool } from "./web-get.js";
import { webSearchTool } from "./web-search.js";

const ALL_NAMES = [
  "web_search",
  "web_get",
  "exec_bash",
  "read_file",
  "read_dir",
  "write_file",
  "merge_file",
] as const;

export function allTools(workspaceRoot: string, tavilyApiKey: string): AgentTool[] {
  return [
    webSearchTool(tavilyApiKey),
    webGetTool(),
    execBashTool(workspaceRoot),
    ...fileOpTools(workspaceRoot),
  ];
}

export function filterTools(
  tools: AgentTool[],
  allowed: string[] | undefined,
): AgentTool[] {
  if (!allowed || allowed.length === 0) {
    return [];
  }
  const set = new Set(allowed);
  return tools.filter((t) => set.has(t.name));
}

export { ALL_NAMES };
