import {
  AcceptNDJSON,
  ContentTypeJSON,
  HeaderSessionID,
  loadAskConfig,
} from "@ai-assistant/core";

export interface AskMessageEntry {
  type: string;
  content: string;
}

export interface AskResult {
  entries: AskMessageEntry[];
  model: string;
  session_id: string;
  tokens: number;
}

export async function runAsk(
  instruction: string,
  opts?: { model?: string; agent?: string },
): Promise<AskResult> {
  const cfg = loadAskConfig();
  const base = cfg.serverURL.replace(/\/$/, "");
  const body = JSON.stringify({
    message: instruction,
    model: opts?.model || cfg.model || undefined,
    agent: opts?.agent,
  });

  const res = await fetch(`${base}/`, {
    method: "POST",
    headers: {
      Accept: AcceptNDJSON,
      "Content-Type": ContentTypeJSON,
    },
    body,
  });

  const sessionId = res.headers.get(HeaderSessionID) || "";
  if (!res.ok) {
    const t = await res.text();
    throw new Error(`server error ${res.status}: ${t}`);
  }

  const entries: AskMessageEntry[] = [];
  let tokenChars = 0;
  const reader = res.body?.getReader();
  if (!reader) {
    throw new Error("no response body");
  }
  const dec = new TextDecoder();
  let buf = "";
  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buf += dec.decode(value, { stream: true });
    let idx: number;
    while ((idx = buf.indexOf("\n")) >= 0) {
      const line = buf.slice(0, idx).trim();
      buf = buf.slice(idx + 1);
      if (!line) {
        continue;
      }
      let j: Record<string, unknown>;
      try {
        j = JSON.parse(line) as Record<string, unknown>;
      } catch {
        continue;
      }
      const typ = j.type as string;
      if (typ === "token" && typeof j.delta === "string") {
        tokenChars += j.delta.length;
        const last = entries[entries.length - 1];
        if (last && last.type === "output") {
          last.content += j.delta;
        } else {
          entries.push({ type: "output", content: j.delta });
        }
      } else if (typ === "thinking" && typeof j.delta === "string") {
        const last = entries[entries.length - 1];
        if (last && last.type === "thinking") {
          last.content += j.delta;
        } else {
          entries.push({ type: "thinking", content: j.delta });
        }
      } else if (typ === "error") {
        throw new Error(String(j.error || "stream error"));
      }
    }
  }

  return {
    entries,
    model: opts?.model || cfg.model || "",
    session_id: sessionId,
    tokens: Math.ceil(tokenChars / 4),
  };
}
