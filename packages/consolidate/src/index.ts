#!/usr/bin/env node
import "dotenv/config";
import { Command } from "commander";
import { resolveWorkspaceRoot } from "@ai-assistant/core";
import { runConsolidation } from "./pipeline.js";

const program = new Command();
program
  .name("ai-assistant-consolidate")
  .option(
    "--workspace <path>",
    "Workspace root (default: AI_ASSISTANT_WORKSPACE or ~/.ai-assistant.workspace)",
  )
  .action(async (opts: { workspace?: string }) => {
    const root = opts.workspace?.trim()
      ? opts.workspace.trim()
      : resolveWorkspaceRoot();
    await runConsolidation(root);
  });

program.parse();
