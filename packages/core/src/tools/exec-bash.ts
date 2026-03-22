import { spawn } from "node:child_process";
import { Type } from "@mariozechner/pi-ai";
import type { AgentTool } from "@mariozechner/pi-agent-core";

const MAX_OUT = 64 * 1024;

export function execBashTool(workspaceRoot: string): AgentTool {
  return {
    name: "exec_bash",
    label: "Run shell command",
    description: "Run a bash command with cwd set to the workspace root.",
    parameters: Type.Object({
      command: Type.String({ description: "Shell command" }),
    }),
    execute: async (_id, params, signal) => {
      const p = params as { command: string };
      return new Promise((resolve) => {
        const child = spawn("/bin/bash", ["-lc", p.command], {
          cwd: workspaceRoot,
          env: { ...process.env, CI: process.env.CI || "" },
        });
        let out = "";
        let err = "";
        const cap = (s: string, add: string) => {
          const next = s + add;
          return next.length > MAX_OUT ? next.slice(0, MAX_OUT) + "\n[truncated]" : next;
        };
        child.stdout?.on("data", (d: Buffer) => {
          out = cap(out, d.toString("utf8"));
        });
        child.stderr?.on("data", (d: Buffer) => {
          err = cap(err, d.toString("utf8"));
        });
        if (signal) {
          signal.addEventListener("abort", () => child.kill("SIGTERM"), {
            once: true,
          });
        }
        child.on("close", (code) => {
          const text = `exit_code: ${code}\nstdout:\n${out}\nstderr:\n${err}`;
          resolve({
            content: [{ type: "text", text }],
            details: { code },
          });
        });
        child.on("error", (e) => {
          resolve({
            content: [{ type: "text", text: `spawn error: ${e}` }],
            details: {},
          });
        });
      });
    },
  };
}
