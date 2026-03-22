import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { buildSystemPrompt } from "./prompt.js";

const bootstrap = {
  includeRing2: true,
  ring2MaxTokens: 500,
  systemPromptMaxTokens: 4096,
};

describe("buildSystemPrompt", () => {
  let root: string;

  beforeEach(async () => {
    root = await fs.mkdtemp(path.join(os.tmpdir(), "ai-prompt-"));
    await fs.writeFile(path.join(root, "SOUL.md"), "soul-line", "utf8");
    await fs.writeFile(path.join(root, "IDENTITY.md"), "id-line", "utf8");
    await fs.writeFile(path.join(root, "USER.md"), "user-fact", "utf8");
    await fs.writeFile(path.join(root, "MEMORY.md"), "mem-fact", "utf8");
    await fs.writeFile(path.join(root, "TASKS.md"), "task-fact", "utf8");
  });

  afterEach(async () => {
    await fs.rm(root, { recursive: true, force: true });
  });

  it("includes rings, agent body, and workspace path", () => {
    const p = buildSystemPrompt(root, "Be helpful.", bootstrap);
    expect(p).toContain("SOUL.md");
    expect(p).toContain("soul-line");
    expect(p).toContain("IDENTITY.md");
    expect(p).toContain("id-line");
    expect(p).toContain("USER.md");
    expect(p).toContain("user-fact");
    expect(p).toContain("Be helpful.");
    expect(p).toContain(root);
  });

  it("omits ring 2 when includeRing2 is false", async () => {
    const p = buildSystemPrompt(root, "x", { ...bootstrap, includeRing2: false });
    expect(p).not.toContain("USER.md");
    expect(p).not.toContain("user-fact");
  });
});
