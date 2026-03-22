import fs from "node:fs";
import path from "node:path";

export function appendTurnLog(
  workspaceRoot: string,
  sessionId: string,
  userMsg: string,
  assistantReply: string,
  turnIndex: number,
  ts: Date = new Date(),
): void {
  const day = ts.toISOString().slice(0, 10);
  const logsDir = path.join(workspaceRoot, "logs");
  fs.mkdirSync(logsDir, { recursive: true });
  const filePath = path.join(logsDir, `${day}-${sessionId}.md`);
  const block = `\n\n---\n<!-- turn=${turnIndex} time=${ts.toISOString()} -->\n\n## User\n\n${userMsg}\n\n## Assistant\n\n${assistantReply}\n`;
  fs.appendFileSync(filePath, block, "utf8");
}
