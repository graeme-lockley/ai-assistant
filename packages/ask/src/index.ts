#!/usr/bin/env node
import "dotenv/config";
import { Command } from "commander";
import { runAsk } from "./ask.js";

const program = new Command();
program
  .name("ai-assistant-ask")
  .argument("<instruction...>", "Instruction to send")
  .option("--model <id>", "Model override")
  .option("--agent <name>", "Agent name (new session only)")
  .action(async (words: string[], opts: { model?: string; agent?: string }) => {
    const instruction = words.join(" ").trim();
    if (!instruction) {
      console.error("instruction required");
      process.exit(1);
    }
    try {
      const result = await runAsk(instruction, {
        model: opts.model,
        agent: opts.agent,
      });
      console.log(JSON.stringify(result));
    } catch (e) {
      console.log(
        JSON.stringify({
          error: e instanceof Error ? e.message : String(e),
          details: "failed to execute ask request",
        }),
      );
      process.exit(1);
    }
  });

program.parse();
