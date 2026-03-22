import { getEnvApiKey, getModel, getModels, type Model } from "@mariozechner/pi-ai";

export type ModelRef =
  | { kind: "builtin"; provider: string; modelId: string }
  | {
      kind: "deepseek";
      modelId: string;
      baseUrl: string;
    };

export function modelIdFromRef(ref: ModelRef): string {
  return ref.modelId;
}

const DEEPSEEK_BASE = "https://api.deepseek.com";

export function deepseekModel(
  modelId: string,
  baseUrl: string = process.env.DEEPSEEK_BASE_URL?.replace(/\/$/, "") ||
    DEEPSEEK_BASE,
): Model<"openai-completions"> {
  return {
    id: modelId,
    name: `Deepseek (${modelId})`,
    api: "openai-completions",
    provider: "deepseek",
    baseUrl,
    reasoning: modelId.includes("reasoner"),
    input: ["text"],
    cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
    contextWindow: 128000,
    maxTokens: 8192,
  };
}

export function modelRefFromModel(m: Model<any>): ModelRef {
  if (m.provider === "deepseek" || m.baseUrl?.includes("deepseek")) {
    return {
      kind: "deepseek",
      modelId: m.id,
      baseUrl: m.baseUrl.replace(/\/$/, "") || DEEPSEEK_BASE,
    };
  }
  return { kind: "builtin", provider: m.provider, modelId: m.id };
}

export function resolveModel(ref: ModelRef): Model<any> {
  if (ref.kind === "deepseek") {
    return deepseekModel(ref.modelId, ref.baseUrl);
  }
  // KnownProvider union — cast modelId for dynamic ids from persisted sessions
  return getModel(
    ref.provider as "anthropic" | "openai",
    ref.modelId as never,
  );
}

export function defaultModelRefFromEnv(): ModelRef {
  const id = process.env.AI_ASSISTANT_DEFAULT_MODEL?.trim();
  const deepseekKey = process.env.DEEPSEEK_API_KEY?.trim();
  const explicitDeepseek =
    id === "deepseek-chat" ||
    id === "deepseek-reasoner" ||
    (!!deepseekKey && !process.env.ANTHROPIC_API_KEY?.trim());

  if (explicitDeepseek && deepseekKey) {
    return {
      kind: "deepseek",
      modelId: id || "deepseek-chat",
      baseUrl:
        process.env.DEEPSEEK_BASE_URL?.replace(/\/$/, "") || DEEPSEEK_BASE,
    };
  }

  if (process.env.ANTHROPIC_API_KEY?.trim()) {
    const mid = id?.startsWith("claude") ? id : "claude-sonnet-4-6";
    return { kind: "builtin", provider: "anthropic", modelId: mid };
  }

  if (process.env.OPENAI_API_KEY?.trim()) {
    const mid = id && !id.startsWith("claude") ? id : "gpt-4o-mini";
    return { kind: "builtin", provider: "openai", modelId: mid };
  }

  if (deepseekKey) {
    return {
      kind: "deepseek",
      modelId: id || "deepseek-chat",
      baseUrl:
        process.env.DEEPSEEK_BASE_URL?.replace(/\/$/, "") || DEEPSEEK_BASE,
    };
  }

  throw new Error(
    "No LLM API key found. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or DEEPSEEK_API_KEY.",
  );
}

export function resolveModelById(modelId: string): Model<any> {
  const trimmed = modelId.trim();
  if (trimmed === "deepseek-chat" || trimmed === "deepseek-reasoner") {
    if (!process.env.DEEPSEEK_API_KEY?.trim()) {
      throw new Error("DEEPSEEK_API_KEY is required for Deepseek models");
    }
    return deepseekModel(trimmed);
  }
  if (trimmed.startsWith("claude")) {
    return getModel("anthropic", trimmed as never);
  }
  return getModel("openai", trimmed as never);
}

const CURATED_ANTHROPIC = [
  "claude-sonnet-4-6",
  "claude-opus-4-6",
  "claude-haiku-4-5",
];
const CURATED_OPENAI = ["gpt-4o-mini", "gpt-4o", "gpt-4.1-mini"];

export function listAvailableModelIds(): string[] {
  const ids = new Set<string>();
  if (process.env.DEEPSEEK_API_KEY?.trim()) {
    ids.add("deepseek-chat");
    ids.add("deepseek-reasoner");
  }
  if (process.env.ANTHROPIC_API_KEY?.trim()) {
    const known = new Set(getModels("anthropic").map((m) => m.id));
    for (const id of CURATED_ANTHROPIC) {
      if (known.has(id)) {
        ids.add(id);
      }
    }
  }
  if (process.env.OPENAI_API_KEY?.trim()) {
    const known = new Set(getModels("openai").map((m) => m.id));
    for (const id of CURATED_OPENAI) {
      if (known.has(id)) {
        ids.add(id);
      }
    }
  }
  return [...ids].sort();
}

export function getApiKeyForModel(model: Model<any>): string | undefined {
  const fromEnv = getEnvApiKey(model.provider);
  if (fromEnv) {
    return fromEnv;
  }
  if (model.provider === "deepseek") {
    return process.env.DEEPSEEK_API_KEY?.trim();
  }
  return undefined;
}
