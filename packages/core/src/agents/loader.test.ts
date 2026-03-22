import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { loadAgentDefinition, listAgentSummaries } from "./loader.js";

describe("agent loader", () => {
  let tmp: string;
  afterEach(() => {
    if (tmp) {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it("loads AGENT.md frontmatter", () => {
    tmp = fs.mkdtempSync(path.join(os.tmpdir(), "aa-"));
    const agents = path.join(tmp, "agents", "test-agent");
    fs.mkdirSync(agents, { recursive: true });
    fs.writeFileSync(
      path.join(agents, "AGENT.md"),
      `---
name: test-agent
description: A test agent for unit tests.
tools:
  - read_file
---
# Body here
Hello
`,
      "utf8",
    );
    const def = loadAgentDefinition(path.join(tmp, "agents"), "test-agent");
    expect(def.name).toBe("test-agent");
    expect(def.description).toContain("unit tests");
    expect(def.tools).toContain("read_file");
    expect(def.body).toContain("Hello");
  });

  it("lists summaries", () => {
    tmp = fs.mkdtempSync(path.join(os.tmpdir(), "aa-"));
    const agents = path.join(tmp, "agents", "a");
    fs.mkdirSync(agents, { recursive: true });
    fs.writeFileSync(
      path.join(agents, "AGENT.md"),
      "---\nname: a\ndescription: desc a\n---\n",
      "utf8",
    );
    const list = listAgentSummaries(tmp);
    expect(list).toHaveLength(1);
    expect(list[0].name).toBe("a");
  });
});
