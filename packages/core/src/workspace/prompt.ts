import path from "node:path";
import type { BootstrapConfig } from "../config.js";
import { readFileIfExists, truncateByTokens } from "./workspace.js";

export interface PromptOptions extends BootstrapConfig {}

/**
 * Assembles the system prompt: Ring 1 (SOUL, IDENTITY), optional Ring 2, agent body.
 */
export function buildSystemPrompt(
  root: string,
  agentBody: string,
  opts: PromptOptions,
): string {
  const minimal = `Priority when sources conflict: SOUL > AGENT (per-agent) > IDENTITY > USER > MEMORY.
Current workspace root: ${root}
`;

  const soul = readFileIfExists(path.join(root, "SOUL.md"));
  const identity = readFileIfExists(path.join(root, "IDENTITY.md"));

  let ring2 = "";
  if (opts.includeRing2) {
    const user = truncateByTokens(
      readFileIfExists(path.join(root, "USER.md")),
      opts.ring2MaxTokens,
    );
    const memory = truncateByTokens(
      readFileIfExists(path.join(root, "MEMORY.md")),
      opts.ring2MaxTokens,
    );
    const tasks = truncateByTokens(
      readFileIfExists(path.join(root, "TASKS.md")),
      opts.ring2MaxTokens,
    );
    ring2 = [
      user && `## USER.md\n${user}`,
      memory && `## MEMORY.md\n${memory}`,
      tasks && `## TASKS.md\n${tasks}`,
    ]
      .filter(Boolean)
      .join("\n\n");
  }

  const agentSection = agentBody.trim()
    ? `## Agent instructions\n${agentBody.trim()}`
    : "";

  let full = [minimal, "## SOUL.md\n" + soul, "## IDENTITY.md\n" + identity]
    .filter((s) => s.trim().length > 0)
    .join("\n\n");
  if (ring2) {
    full += "\n\n" + ring2;
  }
  if (agentSection) {
    full += "\n\n" + agentSection;
  }

  if (opts.systemPromptMaxTokens > 0) {
    full = truncateByTokens(full, opts.systemPromptMaxTokens);
    if (full.includes("[... truncated]")) {
      full += "\n\n[... system prompt truncated]";
    }
  }

  return full.trim();
}
