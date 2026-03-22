import { describe, expect, it } from "vitest";
import { allTools, filterTools } from "./index.js";

describe("filterTools", () => {
  it("returns empty when allowed is empty or undefined", () => {
    const tools = allTools("/tmp/ws", "key");
    expect(filterTools(tools, undefined)).toEqual([]);
    expect(filterTools(tools, [])).toEqual([]);
  });

  it("keeps only tools whose names are listed", () => {
    const tools = allTools("/tmp/ws", "key");
    const names = new Set(tools.map((t) => t.name));
    const subset = filterTools(tools, ["web_search", "read_file"]);
    expect(subset).toHaveLength(2);
    expect(subset.every((t) => names.has(t.name))).toBe(true);
    expect(subset.map((t) => t.name).sort()).toEqual(["read_file", "web_search"]);
  });
});
