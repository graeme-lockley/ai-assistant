import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { buildIndex, indexPath } from "./indexer.js";

describe("indexer", () => {
  let tmp: string;
  afterEach(() => {
    if (tmp) {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it("writes fragments.jsonl", () => {
    tmp = fs.mkdtempSync(path.join(os.tmpdir(), "aa-"));
    fs.mkdirSync(path.join(tmp, "context", "indexes"), { recursive: true });
    fs.writeFileSync(
      path.join(tmp, "SOUL.md"),
      "## One\n\nalpha\n\n## Two\n\nbeta\n",
      "utf8",
    );
    buildIndex(tmp);
    const p = indexPath(tmp);
    expect(fs.existsSync(p)).toBe(true);
    const lines = fs.readFileSync(p, "utf8").trim().split("\n");
    expect(lines.length).toBeGreaterThanOrEqual(1);
    const row = JSON.parse(lines[0] as string) as { fragment_type: string };
    expect(row.fragment_type).toBe("soul");
  });
});
