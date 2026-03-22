import { completeSimple, type Context, type Message } from "@mariozechner/pi-ai";
import {
  defaultModelRefFromEnv,
  ensureWorkspace,
  getApiKeyForModel,
  modelIdFromRef,
  resolveModel,
} from "@ai-assistant/core";
import {
  buildConsolidationUserPrompt,
  collectDailyBundle,
  consolidationSystemPrompt,
  writeWeeklySummaryAndMemory,
  yesterdayISODate,
} from "./stages.js";

export async function runConsolidation(workspaceRoot: string): Promise<void> {
  await ensureWorkspace(workspaceRoot);
  const dayStr = yesterdayISODate();
  const bundle = collectDailyBundle(workspaceRoot, dayStr);
  if (!bundle.trim()) {
    console.log("[consolidate] nothing to consolidate for", dayStr);
    return;
  }

  const ref = defaultModelRefFromEnv();
  const model = resolveModel(ref);
  const apiKey = getApiKeyForModel(model);
  const user = buildConsolidationUserPrompt(dayStr, bundle);

  const messages: Message[] = [
    { role: "user", content: user, timestamp: Date.now() },
  ];
  const ctx: Context = { systemPrompt: consolidationSystemPrompt, messages };

  const reply = await completeSimple(model, ctx, {
    apiKey,
  });

  let text = "";
  for (const block of reply.content) {
    if (block.type === "text") {
      text += block.text;
    }
  }

  const { weeklyPath } = writeWeeklySummaryAndMemory(workspaceRoot, dayStr, text);

  console.log(
    "[consolidate] wrote",
    weeklyPath,
    "and updated MEMORY.md using",
    modelIdFromRef(ref),
  );
}
