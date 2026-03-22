import fs from "node:fs";
import path from "node:path";

/** ISO date string (YYYY-MM-DD) for the calendar day before today (UTC). */
export function yesterdayISODate(): string {
  const d = new Date();
  d.setDate(d.getDate() - 1);
  return d.toISOString().slice(0, 10);
}

/**
 * Gathers yesterday's daily note (if any) and same-day session log snippets under logs/.
 */
export function collectDailyBundle(workspaceRoot: string, dayStr: string): string {
  const dailyDir = path.join(workspaceRoot, "memory", "daily");
  const dailyPath = path.join(dailyDir, `${dayStr}.md`);
  const logsDir = path.join(workspaceRoot, "logs");
  let bundle = "";
  if (fs.existsSync(dailyPath)) {
    bundle += fs.readFileSync(dailyPath, "utf8") + "\n\n";
  }
  if (fs.existsSync(logsDir)) {
    for (const f of fs.readdirSync(logsDir)) {
      if (f.startsWith(dayStr) && f.endsWith(".md")) {
        bundle += fs.readFileSync(path.join(logsDir, f), "utf8") + "\n\n";
      }
    }
  }
  return bundle;
}

export const consolidationSystemPrompt = `You merge short-term notes and session logs into a concise long-term MEMORY.md update.
Output ONLY markdown sections to append or replace under ## Facts, ## Preferences, ## Active Threads, ## Open Questions.
Do not repeat entire old MEMORY; emit deltas only as bullet lists.`;

export function buildConsolidationUserPrompt(dayStr: string, bundle: string): string {
  return `Short-term content for ${dayStr}:\n\n${bundle.slice(0, 120_000)}`;
}

export function writeWeeklySummaryAndMemory(
  workspaceRoot: string,
  dayStr: string,
  summaryText: string,
): { weeklyPath: string; memoryPath: string } {
  const weeklyDir = path.join(workspaceRoot, "memory", "weekly");
  fs.mkdirSync(weeklyDir, { recursive: true });
  const weekPath = path.join(weeklyDir, `${dayStr}-summary.md`);
  fs.writeFileSync(weekPath, `# Week rollup ${dayStr}\n\n${summaryText}\n`, "utf8");

  const memoryPath = path.join(workspaceRoot, "MEMORY.md");
  let mem = fs.existsSync(memoryPath)
    ? fs.readFileSync(memoryPath, "utf8")
    : "# MEMORY\n";
  mem += `\n\n## Consolidated ${dayStr}\n\n${summaryText}\n`;
  fs.writeFileSync(memoryPath, mem, "utf8");
  return { weeklyPath: weekPath, memoryPath };
}
