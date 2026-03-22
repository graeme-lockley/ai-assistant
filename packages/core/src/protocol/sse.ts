import type { Writable } from "node:stream";

function sseLine(field: string, value: string): string {
  const escaped = value.replace(/\r\n/g, "\n").replace(/\n/g, "\\n");
  return `${field}: ${escaped}\n`;
}

export class SSEWriter {
  constructor(private readonly out: { write: (s: string) => void }) {}

  writeEvent(event: string, data?: Record<string, unknown> | null): void {
    let chunk = sseLine("event", event);
    if (data !== undefined && data !== null) {
      chunk += sseLine("data", JSON.stringify(data));
    }
    chunk += "\n";
    this.out.write(chunk);
  }
}

export function sseWriterForNodeStream(stream: Writable): SSEWriter {
  return new SSEWriter({
    write: (s: string) => {
      stream.write(s);
    },
  });
}
