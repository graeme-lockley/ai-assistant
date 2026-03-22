import { describe, expect, it } from "vitest";
import { parseChatRequestBody, ContentTypeJSON, ContentTypePlain } from "./parse.js";

describe("parseChatRequestBody", () => {
  it("parses JSON", () => {
    const r = parseChatRequestBody(
      JSON.stringify({ message: "hi", agent: "assistant" }),
      ContentTypeJSON,
    );
    expect(r.message).toBe("hi");
    expect(r.agent).toBe("assistant");
  });
  it("parses plain text", () => {
    const r = parseChatRequestBody("hello", ContentTypePlain);
    expect(r.message).toBe("hello");
  });
});
