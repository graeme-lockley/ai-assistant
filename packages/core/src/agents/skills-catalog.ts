import fs from "node:fs";
import path from "node:path";
import yaml from "js-yaml";

export function buildSkillsCatalogSection(
  root: string,
  skillNames: string[],
): string {
  if (skillNames.length === 0) {
    return "";
  }
  const lines: string[] = ["## Available skills (activate by reading SKILL.md when relevant)"];
  for (const name of skillNames) {
    const skillPath = path.join(root, "skills", name, "SKILL.md");
    if (!fs.existsSync(skillPath)) {
      continue;
    }
    const raw = fs.readFileSync(skillPath, "utf8");
    const { front } = splitFrontmatter(raw);
    const fm = (yaml.load(front) as { name?: string; description?: string }) || {};
    const n = fm.name || name;
    const d = fm.description || "";
    lines.push(`- **${n}**: ${d} (path: skills/${name}/SKILL.md)`);
  }
  return lines.join("\n");
}

function splitFrontmatter(raw: string): { front: string; body: string } {
  const lines = raw.split(/\r?\n/);
  if (lines[0]?.trim() !== "---") {
    return { front: "", body: raw };
  }
  let i = 1;
  while (i < lines.length && lines[i]?.trim() !== "---") {
    i++;
  }
  if (i >= lines.length) {
    return { front: "", body: raw };
  }
  return {
    front: lines.slice(1, i).join("\n"),
    body: lines.slice(i + 1).join("\n"),
  };
}
