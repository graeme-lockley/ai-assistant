import fs from "node:fs";
import path from "node:path";
import { Type } from "@mariozechner/pi-ai";
import type { AgentTool } from "@mariozechner/pi-agent-core";

function resolveSafe(root: string, rel: string): string {
  const cleaned = path.normalize(rel).replace(/^(\.\.(\/|\\|$))+/, "");
  const abs = path.resolve(root, cleaned);
  const relToRoot = path.relative(root, abs);
  if (relToRoot.startsWith("..") || path.isAbsolute(relToRoot)) {
    throw new Error("path outside workspace root");
  }
  return abs;
}

export function fileOpTools(root: string): AgentTool[] {
  const readFile: AgentTool = {
    name: "read_file",
    label: "Read file",
    description: "Read a UTF-8 text file under the workspace root.",
    parameters: Type.Object({
      path: Type.String({ description: "Relative path from workspace root" }),
    }),
    execute: async (_id, params) => {
      const p = params as { path: string };
      const abs = resolveSafe(root, p.path);
      const text = fs.readFileSync(abs, "utf8");
      return {
        content: [{ type: "text", text }],
        details: { path: p.path },
      };
    },
  };

  const readDir: AgentTool = {
    name: "read_dir",
    label: "List directory",
    description: "List entries in a directory under the workspace root.",
    parameters: Type.Object({
      path: Type.String({ description: "Relative path from workspace root" }),
    }),
    execute: async (_id, params) => {
      const p = params as { path?: string };
      const abs = resolveSafe(root, p.path || ".");
      const names = fs.readdirSync(abs);
      return {
        content: [{ type: "text", text: names.join("\n") }],
        details: { path: p.path },
      };
    },
  };

  const writeFile: AgentTool = {
    name: "write_file",
    label: "Write file",
    description: "Create or overwrite a file under the workspace root.",
    parameters: Type.Object({
      path: Type.String({ description: "Relative path" }),
      content: Type.String({ description: "File contents" }),
    }),
    execute: async (_id, params) => {
      const p = params as { path: string; content: string };
      const abs = resolveSafe(root, p.path);
      fs.mkdirSync(path.dirname(abs), { recursive: true });
      fs.writeFileSync(abs, p.content, "utf8");
      return {
        content: [{ type: "text", text: `wrote ${p.path}` }],
        details: { path: p.path },
      };
    },
  };

  const mergeFile: AgentTool = {
    name: "merge_file",
    label: "Merge into file",
    description:
      "Replace a region in a file: use strategy 'replace' with startLine/endLine (1-based inclusive) or 'markers' with begin_marker/end_marker lines.",
    parameters: Type.Object({
      path: Type.String(),
      content: Type.String(),
      strategy: Type.Union([Type.Literal("replace"), Type.Literal("markers")]),
      start_line: Type.Optional(Type.Number()),
      end_line: Type.Optional(Type.Number()),
      begin_marker: Type.Optional(Type.String()),
      end_marker: Type.Optional(Type.String()),
    }),
    execute: async (_id, params) => {
      const p = params as {
        path: string;
        content: string;
        strategy: "replace" | "markers";
        start_line?: number;
        end_line?: number;
        begin_marker?: string;
        end_marker?: string;
      };
      const abs = resolveSafe(root, p.path);
      let body = fs.readFileSync(abs, "utf8");
      const lines = body.split(/\r?\n/);
      if (p.strategy === "replace") {
        const start = p.start_line ?? 1;
        const end = p.end_line ?? lines.length;
        const before = lines.slice(0, start - 1);
        const after = lines.slice(end);
        const newLines = p.content.split(/\r?\n/);
        body = [...before, ...newLines, ...after].join("\n");
      } else {
        const begin = p.begin_marker;
        const endM = p.end_marker;
        if (!begin || !endM) {
          throw new Error("markers strategy requires begin_marker and end_marker");
        }
        const startIdx = lines.findIndex((l) => l.includes(begin));
        const endIdx = lines.findIndex((l, i) => i > startIdx && l.includes(endM));
        if (startIdx < 0 || endIdx < 0) {
          throw new Error("markers not found");
        }
        const before = lines.slice(0, startIdx + 1);
        const after = lines.slice(endIdx);
        const mid = p.content.split(/\r?\n/);
        body = [...before, ...mid, ...after].join("\n");
      }
      fs.writeFileSync(abs, body, "utf8");
      return {
        content: [{ type: "text", text: `merged ${p.path}` }],
        details: { path: p.path },
      };
    },
  };

  return [readFile, readDir, writeFile, mergeFile];
}
