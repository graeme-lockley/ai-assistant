#!/usr/bin/env node
import "dotenv/config";
import { serve } from "@hono/node-server";
import { Command } from "commander";
import { loadServerConfig } from "@ai-assistant/core";
import { createApp } from "./server.js";

function parseBind(addr: string): { hostname: string; port: number } {
  const a = addr.trim();
  if (a.startsWith(":")) {
    return { hostname: "0.0.0.0", port: parseInt(a.slice(1), 10) || 8080 };
  }
  const idx = a.lastIndexOf(":");
  if (idx <= 0) {
    return { hostname: "0.0.0.0", port: 8080 };
  }
  return {
    hostname: a.slice(0, idx).replace(/^\[/, "").replace(/\]$/, "") || "0.0.0.0",
    port: parseInt(a.slice(idx + 1), 10) || 8080,
  };
}

const program = new Command();
program
  .name("ai-assistant-server")
  .description("HTTP gateway for ai-assistant")
  .action(() => {
    const cfg = loadServerConfig();
    const { hostname, port } = parseBind(cfg.bindAddr);
    const app = createApp();
    serve({ fetch: app.fetch, port, hostname }, (info) => {
      console.log(
        `[server] listening on http://${info.address === "::" ? hostname : info.address}:${info.port}`,
      );
    });
  });

program.parse();
