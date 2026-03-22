#!/usr/bin/env node
import "dotenv/config";
import { Command } from "commander";
import { runRepl } from "./repl.js";

const program = new Command();
program
  .name("ai-assistant-repl")
  .description("TUI REPL client (pi-tui)")
  .action(async () => {
    await runRepl();
  });

program.parse();
