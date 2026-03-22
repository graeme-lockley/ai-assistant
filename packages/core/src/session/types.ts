import type { AgentMessage } from "@mariozechner/pi-agent-core";
import type { ModelRef } from "../model-registry.js";

export interface SessionMetadata {
  id: string;
  agentName: string;
  modelId: string;
  createdAt: string;
  updatedAt: string;
  title?: string;
}

export interface PersistedSessionV1 {
  version: 1;
  meta: SessionMetadata;
  modelRef: ModelRef;
  thinkingLevel: "off" | "minimal" | "low" | "medium" | "high" | "xhigh";
  systemPrompt: string;
  messages: AgentMessage[];
}

export interface SessionStorage {
  save(data: PersistedSessionV1): Promise<void>;
  load(sessionId: string): Promise<PersistedSessionV1 | null>;
  listMeta(): Promise<SessionMetadata[]>;
  delete(sessionId: string): Promise<void>;
}
