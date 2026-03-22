import {
  AcceptNDJSON,
  ContentTypeJSON,
  HeaderSessionClose,
  HeaderSessionID,
  loadReplConfig,
} from "@ai-assistant/core";
import {
  Container,
  Input,
  ProcessTerminal,
  Text,
  TUI,
} from "@mariozechner/pi-tui";

export async function runRepl(): Promise<void> {
  const cfg = loadReplConfig();
  const base = cfg.serverURL.replace(/\/$/, "");
  let sessionId = "";
  let selectedAgent = "assistant";
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

  const append = (s: string) => {
    outBuf += s;
    output.setText(outBuf);
    tui.invalidate();
  };

  async function chatLine(line: string): Promise<void> {
    const res = await fetch(`${base}/`, {
      method: "POST",
      headers: {
        Accept: AcceptNDJSON,
        "Content-Type": ContentTypeJSON,
        ...(sessionId ? { [HeaderSessionID]: sessionId } : {}),
      },
      body: JSON.stringify({
        message: line,
        agent: sessionId ? undefined : selectedAgent,
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
    const reader = res.body?.getReader();
    if (!reader) {
      return;
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
        const ln = buf.slice(0, idx).trim();
        buf = buf.slice(idx + 1);
        if (!ln) {
          continue;
        }
        let j: Record<string, unknown>;
        try {
          j = JSON.parse(ln) as Record<string, unknown>;
        } catch {
          continue;
        }
        if (j.type === "token" && typeof j.delta === "string") {
          append(j.delta);
        } else if (j.type === "error") {
          append(`\n[error] ${String(j.error)}\n`);
        }
      }
    }
    append("\n");
  }

  // pi-tui only delivers keys to focusedComponent; without this, typing does nothing.
  tui.setFocus(input);

  input.onSubmit = async (value: string) => {
    const line = value.trim();
    input.setValue("");
    if (!line) {
      return;
    }
    if (line === "/exit") {
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
      append("\n/help /exit /agents /models /model [id] /agent <name>\n");
      return;
    }
    if (line === "/agents") {
      const r = await fetch(`${base}/agents`);
      append(`\n${await r.text()}\n`);
      return;
    }
    if (line === "/models") {
      const r = await fetch(`${base}/models`);
      append(`\n${await r.text()}\n`);
      return;
    }
    if (line.startsWith("/model ")) {
      const id = line.slice(7).trim();
      if (!sessionId) {
        append("\n[no session yet — send a message first]\n");
        return;
      }
      const r = await fetch(`${base}/model`, {
        method: "POST",
        headers: {
          "Content-Type": ContentTypeJSON,
          [HeaderSessionID]: sessionId,
        },
        body: JSON.stringify({ model: id }),
      });
      append(`\n${await r.text()}\n`);
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
