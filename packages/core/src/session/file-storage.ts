import fs from "node:fs/promises";
import path from "node:path";
import type { PersistedSessionV1, SessionMetadata, SessionStorage } from "./types.js";

export class FileSessionStorage implements SessionStorage {
  constructor(private readonly workspaceRoot: string) {}

  private dir(id: string): string {
    return path.join(this.workspaceRoot, "sessions", id);
  }

  private file(id: string): string {
    return path.join(this.dir(id), "state.json");
  }

  async save(data: PersistedSessionV1): Promise<void> {
    const d = this.dir(data.meta.id);
    await fs.mkdir(d, { recursive: true });
    await fs.writeFile(this.file(data.meta.id), JSON.stringify(data, null, 2), "utf8");
  }

  async load(sessionId: string): Promise<PersistedSessionV1 | null> {
    try {
      const raw = await fs.readFile(this.file(sessionId), "utf8");
      return JSON.parse(raw) as PersistedSessionV1;
    } catch {
      return null;
    }
  }

  async listMeta(): Promise<SessionMetadata[]> {
    const base = path.join(this.workspaceRoot, "sessions");
    let entries: string[] = [];
    try {
      entries = await fs.readdir(base);
    } catch {
      return [];
    }
    const out: SessionMetadata[] = [];
    for (const id of entries) {
      const loaded = await this.load(id);
      if (loaded) {
        out.push(loaded.meta);
      }
    }
    return out.sort(
      (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
    );
  }

  async delete(sessionId: string): Promise<void> {
    await fs.rm(this.dir(sessionId), { recursive: true, force: true });
  }
}
