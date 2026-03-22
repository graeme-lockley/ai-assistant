import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** Resolved path to bundled workspace-template (next to dist/ in dev: ../workspace-template) */
export function templateRoot(): string {
  const fromDist = path.join(__dirname, "..", "workspace-template");
  if (fs.existsSync(fromDist)) {
    return fromDist;
  }
  return path.join(__dirname, "..", "..", "workspace-template");
}

export const REQUIRED_DIRS = [
  "logs",
  "memory/daily",
  "memory/weekly",
  "memory/monthly",
  "skills",
  "tools",
  "agents",
  "sessions",
  "context",
  "context/indexes",
  "context/routing",
  "context/cache",
];

export async function ensureWorkspace(root: string): Promise<void> {
  const abs = path.resolve(root);
  const tpl = templateRoot();
  if (!fs.existsSync(abs)) {
    fs.mkdirSync(abs, { recursive: true });
    copyTemplateIfMissing(tpl, abs);
  } else {
    for (const d of REQUIRED_DIRS) {
      const p = path.join(abs, d);
      if (!fs.existsSync(p)) {
        fs.mkdirSync(p, { recursive: true });
      }
    }
    copyTemplateIfMissing(tpl, abs);
  }
}

function copyTemplateIfMissing(tplRoot: string, destRoot: string): void {
  if (!fs.existsSync(tplRoot)) {
    return;
  }
  walkCopy(tplRoot, tplRoot, destRoot);
}

function walkCopy(tplRoot: string, current: string, destRoot: string): void {
  const entries = fs.readdirSync(current, { withFileTypes: true });
  for (const ent of entries) {
    const src = path.join(current, ent.name);
    const rel = path.relative(tplRoot, src);
    const dest = path.join(destRoot, rel);
    if (ent.isDirectory()) {
      if (!fs.existsSync(dest)) {
        fs.mkdirSync(dest, { recursive: true });
      }
      walkCopy(tplRoot, src, destRoot);
    } else if (!fs.existsSync(dest)) {
      fs.mkdirSync(path.dirname(dest), { recursive: true });
      fs.copyFileSync(src, dest);
    }
  }
}

export function readFileIfExists(p: string): string {
  try {
    return fs.readFileSync(p, "utf8");
  } catch {
    return "";
  }
}

export function truncateByTokens(text: string, maxTokens: number): string {
  if (maxTokens <= 0 || !text) {
    return text;
  }
  const approxCharsPerToken = 4;
  const maxChars = maxTokens * approxCharsPerToken;
  if (text.length <= maxChars) {
    return text;
  }
  return text.slice(0, maxChars) + "\n\n[... truncated]";
}
