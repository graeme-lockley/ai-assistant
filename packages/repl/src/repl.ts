import {
  AcceptNDJSON,
  ContentTypeJSON,
  HeaderSessionClose,
  HeaderSessionID,
  HeaderStreamFormat,
  StreamFormatNDJSON,
  loadReplConfig,
} from "@ai-assistant/core";
import chalk from "chalk";
import type { Component, OverlayHandle, SelectItem, SelectListTheme } from "@mariozechner/pi-tui";
import {
  Box,
  Container,
  Input,
  ProcessTerminal,
  SelectList,
  Text,
  TUI,
} from "@mariozechner/pi-tui";

/** Braille dot spinner (common in CLIs). */
const SPINNER_FRAMES = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

function applyStreamEvent(
  j: Record<string, unknown>,
  onToken: (s: string) => void,
  onErr: (s: string) => void,
): void {
  if (j.type === "token" && typeof j.delta === "string") {
    onToken(j.delta);
  } else if (j.type === "error") {
    onErr(String(j.error));
  }
}

function flushNdjsonLines(
  buf: string,
  onObj: (j: Record<string, unknown>) => void,
): string {
  let idx: number;
  while ((idx = buf.indexOf("\n")) >= 0) {
    const ln = buf.slice(0, idx).trim();
    buf = buf.slice(idx + 1);
    if (!ln) {
      continue;
    }
    try {
      onObj(JSON.parse(ln) as Record<string, unknown>);
    } catch {
      /* skip non-JSON lines */
    }
  }
  return buf;
}

function parseTrailingNdjson(
  buf: string,
  onObj: (j: Record<string, unknown>) => void,
): void {
  const t = buf.trim();
  if (!t) {
    return;
  }
  try {
    onObj(JSON.parse(t) as Record<string, unknown>);
  } catch {
    /* ignore */
  }
}

/** Recover tokens when the server used SSE (`event:` / `data:`) instead of NDJSON. */
function recoverTokensFromSse(
  fullRaw: string,
  onToken: (s: string) => void,
  onErr: (s: string) => void,
): void {
  for (const block of fullRaw.split("\n\n")) {
    let ev = "";
    let data = "";
    for (const line of block.split("\n")) {
      const s = line.trimEnd();
      if (s.startsWith("event:")) {
        ev = s.slice(6).trim();
      } else if (s.startsWith("data:")) {
        data = s.slice(5).trim();
      }
    }
    if (ev === "token" && data) {
      try {
        const j = JSON.parse(data) as { delta?: string };
        if (typeof j.delta === "string") {
          onToken(j.delta);
        }
      } catch {
        /* ignore */
      }
    } else if (ev === "error" && data) {
      try {
        const j = JSON.parse(data) as { error?: string };
        onErr(String(j.error ?? data));
      } catch {
        onErr(data);
      }
    }
  }
}

const MODEL_SELECT_THEME: SelectListTheme = {
  selectedPrefix: (t) => t,
  selectedText: (t) => chalk.cyan.bold(t),
  description: (t) => chalk.dim(t),
  scrollInfo: (t) => chalk.dim(t),
  noMatch: () => chalk.yellow("  No models"),
};

function isAbortError(e: unknown): boolean {
  if (e instanceof Error && e.name === "AbortError") {
    return true;
  }
  if (typeof DOMException !== "undefined" && e instanceof DOMException) {
    return e.name === "AbortError";
  }
  return false;
}

/**
 * Overlay root that forwards keys to SelectList (Box alone has no handleInput).
 */
class ModelPickerOverlay implements Component {
  private readonly box: Box;
  readonly list: SelectList;

  constructor(
    summaryText: string,
    items: SelectItem[],
    initialSelectedIndex: number,
    onSelect: (value: string) => void,
    onCancel: () => void,
  ) {
    const inner = new Container();
    inner.addChild(
      new Text(
        " Pick a model — ↑ / ↓ — Enter to apply — Esc leaves unchanged\n" +
          summaryText,
        0,
        0,
      ),
    );
    const maxVis = Math.max(1, Math.min(14, items.length));
    this.list = new SelectList(items, maxVis, MODEL_SELECT_THEME);
    this.list.setSelectedIndex(
      Math.max(0, Math.min(initialSelectedIndex, items.length - 1)),
    );
    this.list.onSelect = (item) => onSelect(item.value);
    this.list.onCancel = onCancel;
    inner.addChild(this.list);
    this.box = new Box(1, 1, (line) => chalk.bgRgb(45, 45, 45)(line));
    this.box.addChild(inner);
  }

  render(width: number): string[] {
    return this.box.render(width);
  }

  handleInput(data: string): void {
    this.list.handleInput(data);
  }

  invalidate(): void {
    this.box.invalidate();
  }
}

export async function runRepl(): Promise<void> {
  const cfg = loadReplConfig();
  const base = cfg.serverURL.replace(/\/$/, "");
  let sessionId = "";
  let selectedAgent = "assistant";
  let preferredModel = "";
  let outBuf =
    `ai-assistant REPL (pi-tui). Type a message. /help for commands.\nServer: ${base}\n(Run npm run server in another terminal if chat fails.)\n\n`;

  const term = new ProcessTerminal();
  const tui = new TUI(term);
  const output = new Text(outBuf, 0, 0);
  const input = new Input();
  const root = new Container();
  root.addChild(output);
  root.addChild(input);
  tui.addChild(root);

  let chatAbort: AbortController | null = null;
  let isChatStreaming = false;
  let streamPrefix = "";
  let streamBody = "";
  let endSpinTimer: ReturnType<typeof setInterval> | null = null;
  let endSpinFrame = 0;

  function stopEndSpinner(): void {
    if (endSpinTimer !== null) {
      clearInterval(endSpinTimer);
      endSpinTimer = null;
    }
  }

  function refreshStreamingView(): void {
    const spin = `\x1b[2m ${SPINNER_FRAMES[endSpinFrame]}\x1b[0m`;
    output.setText(streamPrefix + streamBody + spin);
    tui.invalidate();
    tui.requestRender();
  }

  function startEndSpinner(): void {
    stopEndSpinner();
    endSpinFrame = 0;
    endSpinTimer = setInterval(() => {
      endSpinFrame = (endSpinFrame + 1) % SPINNER_FRAMES.length;
      refreshStreamingView();
    }, 90);
    refreshStreamingView();
  }

  function commitStreamToTranscript(): void {
    stopEndSpinner();
    outBuf = streamPrefix + streamBody + "\n";
    streamPrefix = "";
    streamBody = "";
    output.setText(outBuf);
    tui.invalidate();
    tui.requestRender();
  }

  function abandonStreamToTranscript(suffix: string): void {
    stopEndSpinner();
    outBuf = streamPrefix + streamBody + suffix;
    streamPrefix = "";
    streamBody = "";
    output.setText(outBuf);
    tui.invalidate();
    tui.requestRender();
  }

  const append = (s: string) => {
    outBuf += s;
    output.setText(outBuf);
    tui.invalidate();
    tui.requestRender();
  };

  async function applyModelChoice(id: string): Promise<void> {
    preferredModel = id;
    append(`\n[model] ${id} — preference saved for new sessions\n`);
    if (sessionId) {
      const r = await fetch(`${base}/model`, {
        method: "POST",
        headers: {
          "Content-Type": ContentTypeJSON,
          [HeaderSessionID]: sessionId,
        },
        body: JSON.stringify({ model: id }),
      });
      const body = await r.text();
      if (r.ok) {
        append(`[model] current session updated: ${body}\n`);
      } else {
        append(`[warn] could not update session (${r.status}): ${body}\n`);
      }
    }
  }

  async function openModelPicker(ids: string[]): Promise<string | null> {
    let sessionModel = "";
    if (sessionId) {
      const mr = await fetch(`${base}/model`, {
        headers: { [HeaderSessionID]: sessionId },
      });
      if (mr.ok) {
        try {
          const j = (await mr.json()) as { model?: string };
          if (typeof j.model === "string") {
            sessionModel = j.model;
          }
        } catch {
          /* ignore */
        }
      }
    }

    const summaryLines: string[] = [];
    if (sessionId) {
      summaryLines.push(
        sessionModel
          ? ` Selected now (this session): ${sessionModel}`
          : ` Selected now (this session): (unknown)`,
      );
    } else {
      summaryLines.push(
        ` No active session — next chat uses preference or server default`,
      );
    }
    summaryLines.push(
      preferredModel
        ? ` New-chat preference: ${preferredModel}`
        : ` New-chat preference: (server default)`,
    );
    const summaryText = "\n" + summaryLines.join("\n") + "\n\n";

    let startIdx = 0;
    if (sessionId && sessionModel && ids.includes(sessionModel)) {
      startIdx = ids.indexOf(sessionModel);
    } else if (preferredModel && ids.includes(preferredModel)) {
      startIdx = ids.indexOf(preferredModel);
    }

    const items: SelectItem[] = ids.map((id) => {
      const tags: string[] = [];
      if (sessionId && id === sessionModel) {
        tags.push("this session");
      }
      if (preferredModel && id === preferredModel) {
        tags.push("new-chat preference");
      }
      const description = tags.length > 0 ? tags.join(" · ") : undefined;
      return { value: id, label: id, description };
    });

    return new Promise((resolve) => {
      let handle: OverlayHandle;
      const overlayRoot = new ModelPickerOverlay(
        summaryText,
        items,
        startIdx,
        (value) => {
          handle.hide();
          tui.setFocus(input);
          resolve(value);
        },
        () => {
          handle.hide();
          tui.setFocus(input);
          resolve(null);
        },
      );
      handle = tui.showOverlay(overlayRoot, {
        width: "78%",
        minWidth: 36,
        maxHeight: "55%",
        anchor: "center",
      });
    });
  }

  async function chatLine(line: string): Promise<void> {
    const ac = new AbortController();
    chatAbort = ac;
    isChatStreaming = true;
    streamPrefix = "";
    streamBody = "";
    try {
      const res = await fetch(`${base}/`, {
        method: "POST",
        signal: ac.signal,
        headers: {
          Accept: AcceptNDJSON,
          [HeaderStreamFormat]: StreamFormatNDJSON,
          "Content-Type": ContentTypeJSON,
          ...(sessionId ? { [HeaderSessionID]: sessionId } : {}),
        },
        body: JSON.stringify({
          message: line,
          agent: sessionId ? undefined : selectedAgent,
          ...(!sessionId && preferredModel ? { model: preferredModel } : {}),
        }),
      });
      const sid = res.headers.get(HeaderSessionID);
      if (sid) {
        sessionId = sid;
      }
      if (!res.ok) {
        append(`\n[error] ${res.status} ${await res.text()}\n`);
        return;
      }
      append("\nAssistant: ");
      streamPrefix = outBuf;
      streamBody = "";
      startEndSpinner();

      const reader = res.body?.getReader();
      if (!reader) {
        abandonStreamToTranscript("\n[error] empty response body\n");
        return;
      }

      const dec = new TextDecoder();
      let buf = "";
      let fullRaw = "";
      const onStreamObj = (j: Record<string, unknown>) => {
        applyStreamEvent(
          j,
          (d) => {
            streamBody += d;
            refreshStreamingView();
          },
          (msg) => {
            streamBody += `\n[error] ${msg}\n`;
            refreshStreamingView();
          },
        );
      };
      try {
        while (true) {
          let done: boolean;
          let value: Uint8Array | undefined;
          try {
            const chunk = await reader.read();
            done = chunk.done;
            value = chunk.value;
          } catch {
            if (ac.signal.aborted) {
              await reader.cancel().catch(() => {});
              abandonStreamToTranscript("\n[stopped — Esc]\n");
              return;
            }
            throw new Error("stream read failed");
          }
          if (done) {
            break;
          }
          const piece = dec.decode(value!, { stream: true });
          buf += piece;
          fullRaw += piece;
          buf = flushNdjsonLines(buf, onStreamObj);
        }
        const tail = dec.decode();
        if (tail) {
          buf += tail;
          fullRaw += tail;
          buf = flushNdjsonLines(buf, onStreamObj);
        }
        parseTrailingNdjson(buf, onStreamObj);
        if (
          streamBody.trim().length === 0 &&
          (fullRaw.includes("event:") || fullRaw.includes("data:"))
        ) {
          recoverTokensFromSse(fullRaw, (d) => {
            streamBody += d;
            refreshStreamingView();
          }, (msg) => {
            streamBody += `\n[error] ${msg}\n`;
            refreshStreamingView();
          });
        }
      } finally {
        try {
          reader.releaseLock();
        } catch {
          /* already released */
        }
      }
      commitStreamToTranscript();
    } catch (e) {
      stopEndSpinner();
      if (isAbortError(e)) {
        if (streamPrefix) {
          abandonStreamToTranscript("\n[stopped — Esc]\n");
        } else {
          append("\n[stopped — Esc]\n");
        }
        return;
      }
      const msg = e instanceof Error ? e.message : String(e);
      if (streamPrefix) {
        abandonStreamToTranscript(`\n[error] ${msg}\n`);
      } else {
        append(`\n[error] ${msg}\n`);
      }
    } finally {
      stopEndSpinner();
      isChatStreaming = false;
      chatAbort = null;
      streamPrefix = "";
      streamBody = "";
    }
  }

  tui.setFocus(input);

  input.onEscape = () => {
    if (chatAbort && isChatStreaming) {
      chatAbort.abort();
    }
  };

  input.onSubmit = async (value: string) => {
    const line = value.trim();
    input.setValue("");
    if (!line) {
      return;
    }
    if (line === "/exit") {
      stopEndSpinner();
      chatAbort?.abort();
      if (sessionId) {
        await fetch(`${base}/`, {
          method: "POST",
          headers: {
            [HeaderSessionClose]: "true",
            [HeaderSessionID]: sessionId,
          },
          body: "{}",
        });
      }
      tui.stop();
      process.exit(0);
      return;
    }
    if (line === "/help") {
      append(
        "\n/help /exit /agents /models /model … /agent <name>\n" +
          "  /models          popup: pick model (Esc cancel)\n" +
          "  /model           show preference and session model\n" +
          "  /model <id|n>    set model by id or /models index\n" +
          "  /model default   clear preference\n" +
          "  Esc while assistant streams — cancel request\n",
      );
      return;
    }
    if (line === "/agents") {
      const r = await fetch(`${base}/agents`);
      append(`\n${await r.text()}\n`);
      return;
    }
    if (line === "/models") {
      const r = await fetch(`${base}/models`);
      const raw = await r.text();
      if (!r.ok) {
        append(`\n[error] ${r.status} ${raw}\n`);
        return;
      }
      let ids: string[] = [];
      try {
        const parsed = JSON.parse(raw) as unknown;
        if (Array.isArray(parsed) && parsed.every((x) => typeof x === "string")) {
          ids = parsed as string[];
        } else {
          append(`\n${raw}\n`);
          return;
        }
      } catch {
        append(`\n${raw}\n`);
        return;
      }
      if (ids.length === 0) {
        append("\n[models] server returned an empty list\n");
        return;
      }
      const picked = await openModelPicker(ids);
      if (picked !== null) {
        await applyModelChoice(picked);
      }
      return;
    }
    if (line === "/model") {
      append(
        `\nPreferred model for new chats: ${preferredModel || "(server default)"}\n`,
      );
      if (sessionId) {
        const r = await fetch(`${base}/model`, {
          headers: { [HeaderSessionID]: sessionId },
        });
        const t = await r.text();
        try {
          const j = JSON.parse(t) as { model?: string };
          if (typeof j.model === "string") {
            append(`Current session model: ${j.model}\n`);
          } else {
            append(`Current session: ${t}\n`);
          }
        } catch {
          append(`Current session: ${t}\n`);
        }
      }
      append("Use /models to pick interactively, or /model <id>.\n");
      return;
    }
    if (line.startsWith("/model ")) {
      let id = line.slice(7).trim();
      if (!id) {
        append("\n[usage] /model <id> | /model <n> | /model default\n");
        return;
      }
      if (id === "default" || id === "clear") {
        preferredModel = "";
        append("\n[model] preference cleared; new sessions use server default\n");
        if (sessionId) {
          append(
            "[note] active session unchanged; use /models or /model <id> to change it\n",
          );
        }
        return;
      }
      if (/^\d+$/.test(id)) {
        const r = await fetch(`${base}/models`);
        const raw = await r.text();
        if (!r.ok) {
          append(`\n[error] ${r.status} ${raw}\n`);
          return;
        }
        let list: string[] = [];
        try {
          const parsed = JSON.parse(raw) as unknown;
          if (Array.isArray(parsed) && parsed.every((x) => typeof x === "string")) {
            list = parsed as string[];
          }
        } catch {
          append(`\n[error] could not parse /models\n`);
          return;
        }
        const n = parseInt(id, 10);
        if (n < 1 || n > list.length) {
          append(`\n[error] no model #${n}; run /models (${list.length} models)\n`);
          return;
        }
        id = list[n - 1]!;
      }
      await applyModelChoice(id);
      return;
    }
    if (line.startsWith("/agent ")) {
      selectedAgent = line.slice(7).trim() || "assistant";
      sessionId = "";
      append(
        `\n[agent set to ${selectedAgent}; new session on next message]\n`,
      );
      return;
    }
    append(`\nYou: ${line}\n`);
    try {
      await chatLine(line);
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      append(`\n[error] ${msg}\n`);
    }
  };

  tui.start();

  void (async () => {
    const ac = new AbortController();
    const t = setTimeout(() => ac.abort(), 2500);
    try {
      const r = await fetch(`${base}/`, { method: "GET", signal: ac.signal });
      if (!r.ok) {
        append(`\n[warn] server returned ${r.status} on GET /\n`);
      }
    } catch {
      append(
        `\n[warn] cannot reach server at ${base} — start it with: npm run server\n`,
      );
    } finally {
      clearTimeout(t);
    }
  })();
}
