import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { collectDailyBundle } from "./stages.js";

describe("collectDailyBundle", () => {
  let root: string;
  const day = "2025-03-10";

  beforeEach(async () => {
    root = await fs.mkdtemp(path.join(os.tmpdir(), "ai-cons-"));
    await fs.mkdir(path.join(root, "memory", "daily"), { recursive: true });
    await fs.mkdir(path.join(root, "logs"), { recursive: true });
  });

  afterEach(async () => {
    await fs.rm(root, { recursive: true, force: true });
  });

  it("merges daily note and matching log files", async () => {
    await fs.writeFile(
      path.join(root, "memory", "daily", `${day}.md`),
      "daily content",
      "utf8",
    );
    await fs.writeFile(
      path.join(root, "logs", `${day}-session.md`),
      "log content",
      "utf8",
    );
    await fs.writeFile(
      path.join(root, "logs", "other.md"),
      "skip",
      "utf8",
    );
    const bundle = collectDailyBundle(root, day);
    expect(bundle).toContain("daily content");
    expect(bundle).toContain("log content");
    expect(bundle).not.toContain("skip");
  });
});
