import { Agent } from "@mariozechner/pi-agent-core";
import type { AgentMessage } from "@mariozechner/pi-agent-core";
import type { Message } from "@mariozechner/pi-ai";
import { getEnvApiKey } from "@mariozechner/pi-ai";
import type { BootstrapConfig } from "../config.js";
import {
  getApiKeyForModel,
  modelRefFromModel,
  resolveModel,
  resolveModelById,
  type ModelRef,
} from "../model-registry.js";
import { buildSkillsCatalogSection } from "./skills-catalog.js";
import { buildSystemPrompt } from "../workspace/prompt.js";
import type { AgentDefinition } from "./loader.js";
import { allTools, filterTools } from "../tools/index.js";

function convertToLlm(messages: AgentMessage[]): Message[] {
  return messages.filter(
    (m): m is Message =>
      m.role === "user" || m.role === "assistant" || m.role === "toolResult",
  ) as Message[];
}

export function composeSystemPrompt(
  workspaceRoot: string,
  def: AgentDefinition,
  bootstrap: BootstrapConfig,
): string {
  const skillsBlock = buildSkillsCatalogSection(workspaceRoot, def.skills);
  const body =
    def.body + (skillsBlock ? `\n\n${skillsBlock}` : "");
  return buildSystemPrompt(workspaceRoot, body, bootstrap);
}

export function createAgentInstance(opts: {
  workspaceRoot: string;
  tavilyApiKey: string;
  def: AgentDefinition;
  bootstrap: BootstrapConfig;
  modelId: string;
  messages?: AgentMessage[];
}): { agent: Agent; modelRef: ModelRef } {
  const systemPrompt = composeSystemPrompt(
    opts.workspaceRoot,
    opts.def,
    opts.bootstrap,
  );
  const model = resolveModelById(opts.modelId);
  const modelRef = modelRefFromModel(model);
  const full = allTools(opts.workspaceRoot, opts.tavilyApiKey);
  const tools = filterTools(full, opts.def.tools);
  const agent = new Agent({
    initialState: {
      systemPrompt,
      model,
      thinkingLevel: "off",
      tools,
      messages: opts.messages ?? [],
    },
    convertToLlm,
    sessionId: undefined,
    getApiKey: (provider) =>
      getEnvApiKey(provider) ||
      getApiKeyForModel(model) ||
      (provider === "deepseek" ? process.env.DEEPSEEK_API_KEY : undefined),
  });
  return { agent, modelRef };
}

export function rehydrateAgentFromPersisted(opts: {
  workspaceRoot: string;
  tavilyApiKey: string;
  def: AgentDefinition;
  bootstrap: BootstrapConfig;
  modelRef: ModelRef;
  systemPrompt: string;
  messages: AgentMessage[];
}): Agent {
  const model = resolveModel(opts.modelRef);
  const full = allTools(opts.workspaceRoot, opts.tavilyApiKey);
  const tools = filterTools(full, opts.def.tools);
  return new Agent({
    initialState: {
      systemPrompt: opts.systemPrompt,
      model,
      thinkingLevel: "off",
      tools,
      messages: opts.messages,
    },
    convertToLlm,
    getApiKey: (provider) =>
      getEnvApiKey(provider) ||
      getApiKeyForModel(model) ||
      (provider === "deepseek" ? process.env.DEEPSEEK_API_KEY : undefined),
  });
}
