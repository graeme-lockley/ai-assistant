export const HeaderSessionID = "X-Session-Id";
export const HeaderSessionClose = "X-Session-Close";

export const ContentTypeJSON = "application/json";
export const ContentTypePlain = "text/plain";
export const ContentTypeSSE = "text/event-stream";

export const AcceptSSE = "text/event-stream";
export const AcceptNDJSON = "application/json";

export type StreamEventType =
  | "session"
  | "agent"
  | "thinking"
  | "token"
  | "tool"
  | "done"
  | "error";

export interface ChatRequestBody {
  message: string;
  model?: string;
  agent?: string;
}

export function parseChatRequestBody(
  body: string,
  contentType: string,
): ChatRequestBody {
  const ct = contentType.split(";")[0]?.trim().toLowerCase() || "";
  if (ct === ContentTypeJSON || ct === "") {
    let j: Record<string, unknown>;
    try {
      j = JSON.parse(body || "{}") as Record<string, unknown>;
    } catch {
      throw new Error("invalid JSON body");
    }
    const message = typeof j.message === "string" ? j.message : "";
    return {
      message,
      model: typeof j.model === "string" ? j.model : undefined,
      agent: typeof j.agent === "string" ? j.agent : undefined,
    };
  }
  if (ct === ContentTypePlain) {
    return { message: body };
  }
  throw new Error(`unsupported content type: ${contentType}`);
}
