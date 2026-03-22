import type { AgentMessage } from "@mariozechner/pi-agent-core";
import { Hono } from "hono";
import {
  ensureIndex,
  ensureWorkspace,
  HeaderSessionClose,
  HeaderSessionID,
  HeaderStreamFormat,
  StreamFormatNDJSON,
  StreamFormatSSE,
  loadServerConfig,
  NDJSONWriter,
  parseChatRequestBody,
  SSEWriter,
  SessionManager,
  listAgentSummaries,
  listAvailableModelIds,
  appendTurnLog,
  ContentTypeJSON,
  ContentTypeSSE,
  defaultModelRefFromEnv,
  modelIdFromRef,
} from "@ai-assistant/core";

function assistantTextFromMessage(m: AgentMessage): string {
  if (m.role !== "assistant") {
    return "";
  }
  const c = m.content;
  if (typeof c === "string") {
    return c;
  }
  if (!Array.isArray(c)) {
    return "";
  }
  const parts: string[] = [];
  for (const block of c) {
    if (typeof block !== "object" || block === null || !("type" in block)) {
      continue;
    }
    if (block.type === "text" && "text" in block) {
      const t = (block as { text?: string }).text;
      if (typeof t === "string" && t.length > 0) {
        parts.push(t);
      }
    } else if (block.type === "thinking" && "thinking" in block) {
      const t = (block as { thinking?: string }).thinking;
      if (typeof t === "string" && t.length > 0) {
        parts.push(t);
      }
    }
  }
  return parts.join("");
}

/** Last assistant message that has visible text (skips tool-only assistant stubs). */
function lastAssistantPlainText(messages: AgentMessage[]): string {
  for (let i = messages.length - 1; i >= 0; i--) {
    const m = messages[i];
    if (m.role !== "assistant") {
      continue;
    }
    const raw = assistantTextFromMessage(m);
    if (raw.trim().length > 0) {
      return raw;
    }
  }
  return "";
}

function explainEmptyAssistantReply(ent: {
  state: { error?: string; messages: AgentMessage[] };
}): string {
  const st = ent.state;
  if (st.error) {
    return `[error] ${st.error}`;
  }
  for (let i = st.messages.length - 1; i >= 0; i--) {
    const m = st.messages[i];
    if (m.role === "assistant" && "errorMessage" in m) {
      const em = (m as { errorMessage?: string }).errorMessage;
      if (em) {
        return `[error] ${em}`;
      }
    }
    if (m.role === "toolResult") {
      const tr = m as { isError?: boolean; content?: unknown };
      if (tr.isError && Array.isArray(tr.content)) {
        const txt = tr.content
          .map((b) =>
            typeof b === "object" && b !== null && "text" in b
              ? String((b as { text: string }).text)
              : "",
          )
          .join(" ")
          .trim();
        if (txt) {
          return `[tool error] ${txt.slice(0, 2000)}`;
        }
      }
    }
  }
  const lastAsst = [...st.messages].reverse().find((x) => x.role === "assistant");
  if (lastAsst) {
    const snippet = JSON.stringify(lastAsst.content).slice(0, 400);
    return `[no visible text] Last assistant blocks: ${snippet}${snippet.length >= 400 ? "…" : ""}`;
  }
  return "[no assistant message in context]";
}

export function createApp(): Hono {
  const cfg = loadServerConfig();
  const defaultRef = defaultModelRefFromEnv();
  const defaultModelId = modelIdFromRef(defaultRef);

  const manager = new SessionManager(
    cfg.workspaceRoot,
    cfg.tavilyApiKey,
    cfg.bootstrap,
    defaultModelId,
  );

  const app = new Hono();

  app.get("/agents", (c) => {
    void ensureWorkspace(cfg.workspaceRoot);
    ensureIndex(cfg.workspaceRoot);
    const agents = listAgentSummaries(cfg.workspaceRoot);
    return c.json(agents);
  });

  app.get("/models", (c) => {
    return c.json(listAvailableModelIds());
  });

  app.get("/sessions", async (c) => {
    const list = await manager.listMeta();
    return c.json(list);
  });

  app.get("/sessions/:id", async (c) => {
    const id = c.req.param("id");
    const ent = await manager.getOrLoad(id);
    if (!ent) {
      return c.json({ error: "not found" }, 404);
    }
    return c.json({
      meta: ent.meta,
      messages: ent.agent.state.messages,
    });
  });

  app.delete("/sessions/:id", async (c) => {
    const id = c.req.param("id");
    await manager.close(id);
    return c.body(null, 204);
  });

  app.get("/model", async (c) => {
    const sid = c.req.header(HeaderSessionID);
    if (!sid) {
      return c.json({ error: "session required" }, 401);
    }
    const ent = await manager.getOrLoad(sid);
    if (!ent) {
      return c.json({ error: "invalid session" }, 401);
    }
    return c.json({ model: ent.meta.modelId });
  });

  app.post("/model", async (c) => {
    const sid = c.req.header(HeaderSessionID);
    if (!sid) {
      return c.json({ error: "session required" }, 401);
    }
    const ent = await manager.getOrLoad(sid);
    if (!ent) {
      return c.json({ error: "invalid session" }, 401);
    }
    const body = await c.req.json().catch(() => ({}));
    const model = typeof body.model === "string" ? body.model.trim() : "";
    if (!model) {
      return c.json({ error: "model required" }, 400);
    }
    const allowed = new Set(listAvailableModelIds());
    if (!allowed.has(model)) {
      return c.json({ error: "unknown model" }, 400);
    }
    manager.setModel(ent, model);
    await manager.persist(ent);
    return c.json({ model });
  });

  app.all("/", async (c) => {
    if (c.req.method === "GET") {
      return c.text("POST / with JSON { message } for chat", 200);
    }
    if (c.req.method !== "POST") {
      return c.text("method not allowed", 405);
    }

    if (c.req.header(HeaderSessionClose) === "true") {
      const sid = c.req.header(HeaderSessionID);
      if (sid) {
        await manager.close(sid);
      }
      return c.body(null, 204);
    }

    await ensureWorkspace(cfg.workspaceRoot);
    ensureIndex(cfg.workspaceRoot);

    const ct = c.req.header("Content-Type") || ContentTypeJSON;
    const rawBody = await c.req.text();
    let parsed: { message: string; model?: string; agent?: string };
    try {
      parsed = parseChatRequestBody(rawBody, ct);
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      if (msg.includes("unsupported content type")) {
        return c.text(msg, 415);
      }
      return c.text(msg, 400);
    }

    let sessionId = c.req.header(HeaderSessionID) || "";
    let ent = sessionId ? await manager.getOrLoad(sessionId) : null;
    let newSession = false;

    if (!ent) {
      if (sessionId) {
        return c.text("invalid or expired session", 401);
      }
      const agentName = parsed.agent?.trim() || "assistant";
      ent = await manager.create(agentName, parsed.model?.trim());
      newSession = true;
      sessionId = ent.meta.id;
    }

    if (newSession && parsed.model?.trim()) {
      try {
        manager.setModel(ent, parsed.model.trim());
      } catch {
        // ignore invalid model on create
      }
    }

    const fmtHdr = c.req.header(HeaderStreamFormat)?.trim().toLowerCase();
    let streamFormat: "ndjson" | "sse";
    if (fmtHdr === StreamFormatNDJSON) {
      streamFormat = "ndjson";
    } else if (fmtHdr === StreamFormatSSE) {
      streamFormat = "sse";
    } else {
      const accept =
        c.req.header("Accept")?.toLowerCase() || cfg.defaultResponseType;
      const useSSE = accept.includes("event-stream");
      const useNdjson =
        accept.includes("application/json") && !useSSE;
      streamFormat = useNdjson ? "ndjson" : "sse";
    }

    const headers: Record<string, string> = {
      [HeaderSessionID]: sessionId,
      "Cache-Control": "no-cache",
      "X-Accel-Buffering": "no",
    };

    if (streamFormat === "sse") {
      headers["Content-Type"] = ContentTypeSSE;
    } else {
      headers["Content-Type"] = ContentTypeJSON;
    }

    const userMessage = parsed.message;
    if (!userMessage.trim()) {
      return c.text("message required", 400);
    }

    const enc = new TextEncoder();
    let streamCtl: ReadableStreamDefaultController<Uint8Array> | undefined;
    const readable = new ReadableStream<Uint8Array>({
      start(controller) {
        streamCtl = controller;
      },
    });
    const writeStr = (s: string) => {
      if (!streamCtl) {
        return;
      }
      try {
        streamCtl.enqueue(enc.encode(s));
      } catch {
        /* closed */
      }
    };

    const sse = new SSEWriter({ write: writeStr });
    const ndj = new NDJSONWriter({ write: writeStr });

    void (async () => {
      try {
        if (streamFormat === "sse") {
          if (newSession) {
            sse.writeEvent("session", { session_id: sessionId });
            sse.writeEvent("agent", { agent: ent!.meta.agentName });
          }
        } else {
          if (newSession) {
            ndj.writeLine({ type: "session", session_id: sessionId });
            ndj.writeLine({ type: "agent", agent: ent!.meta.agentName });
          }
        }

        let fullReply = "";
        /** Text already sent to the client as `token` deltas (for remainder flush). */
        let streamedTokenAcc = "";
        /** Per pi-ai: some providers emit `text_end` with full segment text but no `text_delta`. */
        let textDeltasSinceTextStart = 0;
        const unsub = ent!.agent.subscribe((ev) => {
          if (ev.type === "message_update") {
            const ame = ev.assistantMessageEvent;
            if (ame.type === "text_start") {
              textDeltasSinceTextStart = 0;
            } else if (ame.type === "text_delta") {
              textDeltasSinceTextStart += 1;
              fullReply += ame.delta;
              streamedTokenAcc += ame.delta;
              if (streamFormat === "sse") {
                sse.writeEvent("token", { delta: ame.delta });
              } else {
                ndj.writeLine({ type: "token", delta: ame.delta });
              }
            } else if (ame.type === "text_end") {
              const seg = ame.content;
              if (
                typeof seg === "string" &&
                seg.length > 0 &&
                textDeltasSinceTextStart === 0
              ) {
                fullReply += seg;
                streamedTokenAcc += seg;
                if (streamFormat === "sse") {
                  sse.writeEvent("token", { delta: seg });
                } else {
                  ndj.writeLine({ type: "token", delta: seg });
                }
              }
            } else if (ame.type === "thinking_delta") {
              if (streamFormat === "sse") {
                sse.writeEvent("thinking", { delta: ame.delta });
              } else {
                ndj.writeLine({ type: "thinking", delta: ame.delta });
              }
            }
          } else if (ev.type === "tool_execution_start") {
            if (streamFormat === "sse") {
              sse.writeEvent("tool", { name: ev.toolName });
            } else {
              ndj.writeLine({ type: "tool", name: ev.toolName });
            }
          }
        });

        try {
          await ent!.agent.prompt(userMessage);
        } finally {
          unsub();
        }

        const finalText = lastAssistantPlainText(ent!.agent.state.messages);
        fullReply = finalText || fullReply;

        // Some providers finish with `done` / `message_end` without `text_delta` chunks.
        // Flush any assistant text that was never streamed as tokens.
        if (fullReply.length > streamedTokenAcc.length) {
          const rest = fullReply.slice(streamedTokenAcc.length);
          if (streamFormat === "sse") {
            sse.writeEvent("token", { delta: rest });
          } else {
            ndj.writeLine({ type: "token", delta: rest });
          }
        } else if (fullReply.trim().length === 0) {
          const explain = explainEmptyAssistantReply(ent!.agent);
          if (streamFormat === "sse") {
            sse.writeEvent("token", { delta: explain + "\n" });
          } else {
            ndj.writeLine({ type: "token", delta: explain + "\n" });
          }
          fullReply = explain;
        }

        if (process.env.AI_ASSISTANT_DEBUG_STREAM === "1") {
          const roles = ent!.agent.state.messages.map((m) => m.role).join(",");
          console.log(
            `[stream] roles=${roles} fullReplyLen=${fullReply.length} streamedLen=${streamedTokenAcc.length}`,
          );
        }

        const turnCount = ent!.agent.state.messages.filter(
          (m) => m.role === "user",
        ).length;
        appendTurnLog(
          cfg.workspaceRoot,
          sessionId,
          userMessage,
          fullReply,
          turnCount,
        );

        await manager.persist(ent!);

        if (streamFormat === "sse") {
          sse.writeEvent("done", null);
        } else {
          ndj.writeLine({ type: "done" });
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        if (streamFormat === "sse") {
          sse.writeEvent("error", { error: msg });
        } else {
          ndj.writeLine({ type: "error", error: msg });
        }
      } finally {
        try {
          streamCtl?.close();
        } catch {
          /* ignore */
        }
      }
    })();

    return new Response(readable, { status: 200, headers });
  });

  return app;
}
