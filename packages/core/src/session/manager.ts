import { v4 as uuidv4 } from "uuid";
import type { Agent } from "@mariozechner/pi-agent-core";
import type { BootstrapConfig } from "../config.js";
import { loadAgentByName } from "../agents/loader.js";
import {
  createAgentInstance,
  rehydrateAgentFromPersisted,
} from "../agents/factory.js";
import { modelRefFromModel, resolveModelById } from "../model-registry.js";
import type { ModelRef } from "../model-registry.js";
import { FileSessionStorage } from "./file-storage.js";
import type { PersistedSessionV1, SessionMetadata } from "./types.js";
export interface SessionEntry {
  agent: Agent;
  meta: SessionMetadata;
  modelRef: ModelRef;
  systemPrompt: string;
}

export class SessionManager {
  private readonly storage: FileSessionStorage;
  private readonly hot = new Map<string, SessionEntry>();

  constructor(
    private readonly workspaceRoot: string,
    private readonly tavilyApiKey: string,
    private readonly bootstrap: BootstrapConfig,
    private readonly defaultModelId: string,
  ) {
    this.storage = new FileSessionStorage(workspaceRoot);
  }

  async create(agentName: string, modelId?: string): Promise<SessionEntry> {
    const def = loadAgentByName(this.workspaceRoot, agentName);
    const mid = modelId?.trim() || def.model?.trim() || this.defaultModelId;
    const { agent, modelRef } = createAgentInstance({
      workspaceRoot: this.workspaceRoot,
      tavilyApiKey: this.tavilyApiKey,
      def,
      bootstrap: this.bootstrap,
      modelId: mid,
    });
    const id = uuidv4();
    const now = new Date().toISOString();
    const meta: SessionMetadata = {
      id,
      agentName: def.name,
      modelId: mid,
      createdAt: now,
      updatedAt: now,
    };
    const systemPrompt = agent.state.systemPrompt;
    const ent: SessionEntry = {
      agent,
      meta,
      modelRef,
      systemPrompt,
    };
    agent.sessionId = id;
    this.hot.set(id, ent);
    await this.persist(ent);
    return ent;
  }

  async getOrLoad(sessionId: string): Promise<SessionEntry | null> {
    const hot = this.hot.get(sessionId);
    if (hot) {
      return hot;
    }
    const data = await this.storage.load(sessionId);
    if (!data) {
      return null;
    }
    const def = loadAgentByName(this.workspaceRoot, data.meta.agentName);
    const agent = rehydrateAgentFromPersisted({
      workspaceRoot: this.workspaceRoot,
      tavilyApiKey: this.tavilyApiKey,
      def,
      bootstrap: this.bootstrap,
      modelRef: data.modelRef,
      systemPrompt: data.systemPrompt,
      messages: data.messages,
    });
    const ent: SessionEntry = {
      agent,
      meta: data.meta,
      modelRef: data.modelRef,
      systemPrompt: data.systemPrompt,
    };
    agent.sessionId = sessionId;
    this.hot.set(sessionId, ent);
    return ent;
  }

  async persist(ent: SessionEntry): Promise<void> {
    const now = new Date().toISOString();
    ent.meta.updatedAt = now;
    const data: PersistedSessionV1 = {
      version: 1,
      meta: ent.meta,
      modelRef: ent.modelRef,
      thinkingLevel: "off",
      systemPrompt: ent.agent.state.systemPrompt,
      messages: ent.agent.state.messages,
    };
    await this.storage.save(data);
  }

  setModel(ent: SessionEntry, modelId: string): void {
    const m = resolveModelById(modelId);
    ent.agent.setModel(m);
    ent.modelRef = modelRefFromModel(m);
    ent.meta.modelId = modelId;
  }

  async close(sessionId: string): Promise<void> {
    this.hot.delete(sessionId);
    await this.storage.delete(sessionId);
  }

  async listMeta(): Promise<SessionMetadata[]> {
    return this.storage.listMeta();
  }

}
