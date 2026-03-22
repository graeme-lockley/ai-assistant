import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { FileSessionStorage } from "./file-storage.js";
import type { PersistedSessionV1 } from "./types.js";

function sampleSession(id: string, updatedAt: string): PersistedSessionV1 {
  return {
    version: 1,
    meta: {
      id,
      agentName: "assistant",
      modelId: "deepseek-chat",
      createdAt: "2025-01-01T00:00:00.000Z",
      updatedAt,
    },
    modelRef: { kind: "builtin", provider: "deepseek", modelId: "deepseek-chat" },
    thinkingLevel: "off",
    systemPrompt: "sys",
    messages: [],
  };
}

describe("FileSessionStorage", () => {
  let tmp: string;
  let storage: FileSessionStorage;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), "ai-sess-"));
    await fs.mkdir(path.join(tmp, "sessions"), { recursive: true });
    storage = new FileSessionStorage(tmp);
  });

  afterEach(async () => {
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it("saves and loads round-trip", async () => {
    const data = sampleSession("abc-123", "2025-01-02T00:00:00.000Z");
    await storage.save(data);
    const loaded = await storage.load("abc-123");
    expect(loaded).toEqual(data);
  });

  it("returns null when session file is missing", async () => {
    expect(await storage.load("missing")).toBeNull();
  });

  it("listMeta sorts by updatedAt descending", async () => {
    await storage.save(
      sampleSession("older", "2025-01-01T00:00:00.000Z"),
    );
    await storage.save(
      sampleSession("newer", "2025-01-03T00:00:00.000Z"),
    );
    const list = await storage.listMeta();
    expect(list.map((m) => m.id)).toEqual(["newer", "older"]);
  });

  it("delete removes the session directory", async () => {
    await storage.save(sampleSession("x", "2025-01-01T00:00:00.000Z"));
    await storage.delete("x");
    expect(await storage.load("x")).toBeNull();
    expect(await storage.listMeta()).toEqual([]);
  });
});
