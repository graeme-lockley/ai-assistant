import fs from "node:fs";
import path from "node:path";
import yaml from "js-yaml";

export interface AgentDefinition {
  name: string;
  description: string;
  tools: string[];
  skills: string[];
  model?: string;
  body: string;
  dir: string;
}

interface Frontmatter {
  name?: string;
  description?: string;
  tools?: string[];
  skills?: string[];
  model?: string;
}

export function listAgentSummaries(root: string): { name: string; description: string }[] {
  const agentsDir = path.join(root, "agents");
  if (!fs.existsSync(agentsDir)) {
    return [];
  }
  const out: { name: string; description: string }[] = [];
  for (const ent of fs.readdirSync(agentsDir, { withFileTypes: true })) {
    if (!ent.isDirectory()) {
      continue;
    }
    const defPath = path.join(agentsDir, ent.name, "AGENT.md");
    if (!fs.existsSync(defPath)) {
      continue;
    }
    try {
      const def = loadAgentDefinition(agentsDir, ent.name);
      out.push({ name: def.name, description: def.description });
    } catch {
      // skip invalid
    }
  }
  return out.sort((a, b) => a.name.localeCompare(b.name));
}

export function loadAgentDefinition(
  agentsDir: string,
  dirName: string,
): AgentDefinition {
  const dir = path.join(agentsDir, dirName);
  const filePath = path.join(dir, "AGENT.md");
  const raw = fs.readFileSync(filePath, "utf8");
  const { front, body } = splitFrontmatter(raw);
  const fm = (yaml.load(front) as Frontmatter) || {};
  const name = (fm.name || dirName).trim().toLowerCase();
  const description = (fm.description || "").trim();
  if (!description) {
    throw new Error(`agent ${dirName}: description is required in frontmatter`);
  }
  return {
    name,
    description,
    tools: Array.isArray(fm.tools) ? fm.tools.map(String) : [],
    skills: Array.isArray(fm.skills) ? fm.skills.map(String) : [],
    model: fm.model?.trim(),
    body: body.trim(),
    dir,
  };
}

export function loadAgentByName(root: string, agentName: string): AgentDefinition {
  const agentsDir = path.join(root, "agents");
  const normalized = agentName.trim().toLowerCase();
  for (const ent of fs.readdirSync(agentsDir, { withFileTypes: true })) {
    if (!ent.isDirectory()) {
      continue;
    }
    if (ent.name === normalized) {
      return loadAgentDefinition(agentsDir, ent.name);
    }
  }
  throw new Error(`unknown agent: ${agentName}`);
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
  const front = lines.slice(1, i).join("\n");
  const body = lines.slice(i + 1).join("\n");
  return { front, body };
}
