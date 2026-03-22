import fs from "node:fs";
import path from "node:path";
import { createHash } from "node:crypto";

export interface Fragment {
  id: string;
  source_file: string;
  source_path: string;
  section_path: string[];
  fragment_type: string;
  content: string;
  tags: string[];
  estimated_tokens: number;
  created_at: string;
  updated_at: string;
}

const SOURCE_FILES: Record<string, string> = {
  "IDENTITY.md": "identity",
  "SOUL.md": "soul",
  "USER.md": "user",
  "MEMORY.md": "memory",
  "TASKS.md": "tasks",
};

function estimateTokens(s: string): number {
  return Math.max(1, Math.ceil(s.length / 4));
}

function hashId(parts: string[]): string {
  const h = createHash("sha256");
  h.update(parts.join("|"));
  return h.digest("hex").slice(0, 16);
}

function parseSections(
  relPath: string,
  body: string,
  fragType: string,
): Fragment[] {
  const lines = body.split(/\r?\n/);
  const frags: Fragment[] = [];
  let sectionPath: string[] = [];
  let chunkLines: string[] = [];
  const flush = (title: string) => {
    const content = chunkLines.join("\n").trim();
    if (!content && title === "") {
      return;
    }
    const sp = [...sectionPath];
    const id = hashId([relPath, ...sp, title]);
    const now = new Date().toISOString();
    frags.push({
      id,
      source_file: path.basename(relPath),
      source_path: relPath,
      section_path: sp,
      fragment_type: fragType,
      content: title ? `## ${title}\n\n${content}` : content,
      tags: [],
      estimated_tokens: estimateTokens(content),
      created_at: now,
      updated_at: now,
    });
  };

  for (const line of lines) {
    const h2 = line.match(/^## (.+)/);
    const h3 = line.match(/^### (.+)/);
    if (h2) {
      flush("");
      sectionPath = [h2[1].trim()];
      chunkLines = [];
    } else if (h3) {
      flush("");
      sectionPath = sectionPath.length ? [sectionPath[0], h3[1].trim()] : [h3[1].trim()];
      chunkLines = [];
    } else {
      chunkLines.push(line);
    }
  }
  flush("");
  return frags;
}

export function indexPath(root: string): string {
  return path.join(root, "context", "indexes", "fragments.jsonl");
}

export function needsRebuild(root: string): boolean {
  const idx = indexPath(root);
  let idxTime = 0;
  try {
    idxTime = fs.statSync(idx).mtimeMs;
  } catch {
    return true;
  }
  for (const f of Object.keys(SOURCE_FILES)) {
    const p = path.join(root, f);
    try {
      if (fs.statSync(p).mtimeMs > idxTime) {
        return true;
      }
    } catch {
      // missing
    }
  }
  const skillsDir = path.join(root, "skills");
  if (fs.existsSync(skillsDir)) {
    for (const ent of fs.readdirSync(skillsDir, { withFileTypes: true })) {
      if (!ent.isDirectory()) {
        continue;
      }
      const sp = path.join(skillsDir, ent.name, "SKILL.md");
      try {
        if (fs.statSync(sp).mtimeMs > idxTime) {
          return true;
        }
      } catch {
        // skip
      }
    }
  }
  return false;
}

export function ensureIndex(root: string): void {
  if (!needsRebuild(root)) {
    return;
  }
  buildIndex(root);
}

export function buildIndex(root: string): void {
  const frags: Fragment[] = [];
  for (const [file, typ] of Object.entries(SOURCE_FILES)) {
    const p = path.join(root, file);
    if (!fs.existsSync(p)) {
      continue;
    }
    const body = fs.readFileSync(p, "utf8");
    frags.push(...parseSections(file, body, typ));
  }
  const skillsDir = path.join(root, "skills");
  if (fs.existsSync(skillsDir)) {
    for (const ent of fs.readdirSync(skillsDir, { withFileTypes: true })) {
      if (!ent.isDirectory()) {
        continue;
      }
      const sp = path.join(skillsDir, ent.name, "SKILL.md");
      if (!fs.existsSync(sp)) {
        continue;
      }
      const rel = path.join("skills", ent.name, "SKILL.md");
      const body = fs.readFileSync(sp, "utf8");
      frags.push(...parseSections(rel, body, "skill"));
    }
  }
  const out = indexPath(root);
  fs.mkdirSync(path.dirname(out), { recursive: true });
  const lines = frags.map((f) => JSON.stringify(f));
  fs.writeFileSync(out, lines.join("\n") + (lines.length ? "\n" : ""), "utf8");
}
