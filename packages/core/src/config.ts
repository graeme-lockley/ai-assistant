import { homedir } from "node:os";
import path from "node:path";

export interface BootstrapConfig {
  includeRing2: boolean;
  ring2MaxTokens: number;
  systemPromptMaxTokens: number;
}

export interface ServerConfig {
  bindAddr: string;
  workspaceRoot: string;
  defaultResponseType: string;
  bootstrap: BootstrapConfig;
  tavilyApiKey: string;
  /** Default model id string (e.g. claude-sonnet-4-6, deepseek-chat, gpt-4o-mini) */
  defaultModelId: string;
}

export interface ReplConfig {
  serverURL: string;
  defaultResponseType: string;
}

export interface AskConfig {
  serverURL: string;
  model: string;
  defaultResponseType: string;
}

function envBool(name: string, defaultValue: boolean): boolean {
  const v = process.env[name];
  if (v === undefined || v === "") {
    return defaultValue;
  }
  return v === "1" || v.toLowerCase() === "true";
}

function envInt(name: string, defaultValue: number): number {
  const v = process.env[name];
  if (v === undefined || v === "") {
    return defaultValue;
  }
  const n = parseInt(v, 10);
  return Number.isFinite(n) ? n : defaultValue;
}

export function resolveWorkspaceRoot(): string {
  const fromEnv =
    process.env.AI_ASSISTANT_WORKSPACE?.trim() ||
    process.env.AI_ASSISTANT_ROOT_DIR?.trim();
  if (fromEnv) {
    return expandHome(fromEnv);
  }
  return path.join(homedir(), ".ai-assistant.workspace");
}

function expandHome(p: string): string {
  if (p === "~" || p.startsWith("~/")) {
    return path.join(homedir(), p.slice(1).replace(/^\//, ""));
  }
  return path.resolve(p);
}

export function loadServerConfig(): ServerConfig {
  const bind =
    process.env.AI_ASSISTANT_BIND?.trim() ||
    process.env.BIND_ADDR?.trim() ||
    ":8080";
  return {
    bindAddr: bind,
    workspaceRoot: resolveWorkspaceRoot(),
    defaultResponseType:
      process.env.AI_ASSISTANT_DEFAULT_RESPONSE_TYPE?.trim() ||
      "text/event-stream",
    bootstrap: {
      includeRing2: envBool("AI_ASSISTANT_BOOTSTRAP_RING2", true),
      ring2MaxTokens: envInt("AI_ASSISTANT_RING2_MAX_TOKENS", 500),
      systemPromptMaxTokens: envInt("AI_ASSISTANT_SYSTEM_PROMPT_MAX_TOKENS", 4096),
    },
    tavilyApiKey: process.env.TAVILY_API_KEY?.trim() || "",
    defaultModelId:
      process.env.AI_ASSISTANT_DEFAULT_MODEL?.trim() ||
      process.env.DEEPSEEK_MODEL?.trim() ||
      "claude-sonnet-4-6",
  };
}

export function defaultServerURL(): string {
  const full = process.env.AI_ASSISTANT_SERVER_URL?.trim();
  if (full) {
    return full.replace(/\/$/, "");
  }
  const addr =
    process.env.AI_ASSISTANT_SERVER_ADDR?.trim() || "127.0.0.1:8080";
  const hostPort = addr.includes("://") ? addr : `http://${addr}`;
  return hostPort.replace(/\/$/, "");
}

export function loadReplConfig(): ReplConfig {
  return {
    serverURL: defaultServerURL(),
    defaultResponseType:
      process.env.AI_ASSISTANT_DEFAULT_RESPONSE_TYPE?.trim() ||
      "text/event-stream",
  };
}

export function loadAskConfig(): AskConfig {
  return {
    serverURL: defaultServerURL(),
    model: process.env.AI_ASSISTANT_ASK_MODEL?.trim() || "",
    defaultResponseType:
      process.env.AI_ASSISTANT_DEFAULT_RESPONSE_TYPE?.trim() ||
      "text/event-stream",
  };
}
